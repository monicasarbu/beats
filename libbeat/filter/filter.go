package filter

import (
	"strings"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

type FilterCondition struct {
}

type DropFieldsConfig struct {
	Fields []string `yaml:"fields"`
}

type IncludeFieldsConfig struct {
	Fields []string `yaml:"fields"`
}

type FilterConfig struct {
	DropFields    *DropFieldsConfig    `yaml:"drop_fields"`
	IncludeFields *IncludeFieldsConfig `yaml:"include_fields"`
}

type FilterRule interface {
	Filter(event *common.MapStr) error
	String() string
}

type IncludeFields struct {
	Fields []string
	// condition
}

type DropFields struct {
	Fields []string
	// condition
}

type FilterList struct {
	filters []FilterRule
}

func New(config []FilterConfig) (*FilterList, error) {

	Filters := &FilterList{}
	Filters.filters = []FilterRule{}

	for i, filterConfig := range config {
		logp.Debug("filter", "drop fields=%v include fields=%v", filterConfig.DropFields, filterConfig.IncludeFields)

		if filterConfig.DropFields != nil {
			Filters.Register(i, &DropFields{Fields: filterConfig.DropFields.Fields})
		}

		if filterConfig.IncludeFields != nil {
			Filters.Register(i, &IncludeFields{Fields: filterConfig.IncludeFields.Fields})
		}
	}

	logp.Debug("filter", "filters: %v", Filters)
	return Filters, nil
}

func (filters *FilterList) Register(index int, filter FilterRule) {
	filters.filters = append(filters.filters, filter)
}

func (filters *FilterList) Get(index int) FilterRule {
	return filters.filters[index]
}

func (filters *FilterList) Filter(event *common.MapStr) error {

	for _, filter := range filters.filters {
		if err := filter.Filter(event); err != nil {
			logp.Err("Failed to filter the event: %v", err)
			return err
		}
	}

	return nil
}

func (filters *FilterList) String() string {
	s := []string{}

	for _, filter := range filters.filters {

		s = append(s, filter.String())
	}
	return strings.Join(s, ", ")
}

func (f *IncludeFields) Filter(event *common.MapStr) error {

	logp.Debug("filter", "call include_fields\n")

	return nil
}
func (f *IncludeFields) String() string {
	return "include_fields=" + strings.Join(f.Fields, ", ")
}

func (f *DropFields) Filter(event *common.MapStr) error {

	logp.Debug("filter", "call drop_fields\n %v\n", *event)

	for _, field := range f.Fields {
		event.Delete(field)

	}
	logp.Debug("filter", "after dropping %v", *event)
	return nil
}

func (f *DropFields) String() string {

	return "drop_fields=" + strings.Join(f.Fields, ", ")
}
