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
	Filter(event *common.MapStr)
	String() string
}

/* extends FilterRule */
type IncludeFields struct {
	Fields []string
	// condition
}

/* extend FilterRule */
type DropFields struct {
	Fields []string
	// condition
}

type FilterList struct {
	filters []FilterRule
}

// fields that should be always exported
var ReadOnlyFields = []string{"@timestamp", "beat", "type"}

/* FilterList methods */
func New(config []FilterConfig) (*FilterList, error) {

	Filters := &FilterList{}
	Filters.filters = []FilterRule{}

	for i, filterConfig := range config {
		if filterConfig.DropFields != nil {
			Filters.Register(i, NewDropFields(filterConfig.DropFields.Fields))
		}

		if filterConfig.IncludeFields != nil {
			Filters.Register(i, NewIncludeFields(filterConfig.IncludeFields.Fields))
		}
	}

	logp.Debug("filter", "filters: %v", Filters)
	return Filters, nil
}

func (filters *FilterList) Register(index int, filter FilterRule) {
	filters.filters = append(filters.filters, filter)
	logp.Debug("filter", "Register filter: %v", filter)
}

func (filters *FilterList) Get(index int) FilterRule {
	return filters.filters[index]
}

func (filters *FilterList) Filter(event *common.MapStr) {

	for _, filter := range filters.filters {
		filter.Filter(event)
	}

}

func (filters *FilterList) String() string {
	s := []string{}

	for _, filter := range filters.filters {

		s = append(s, filter.String())
	}
	return strings.Join(s, ", ")
}

/* IncludeFields methods */
func NewIncludeFields(fields []string) *IncludeFields {

	/* add read only fields if they are not yet */
	for _, readOnly := range ReadOnlyFields {
		found := false
		for _, field := range fields {
			if readOnly == field {
				found = true
			}
		}
		if !found {
			fields = append(fields, readOnly)
		}
	}
	return &IncludeFields{Fields: fields}
}

func (f *IncludeFields) Filter(event *common.MapStr) {

	newEvent := &common.MapStr{}

	for _, field := range f.Fields {
		newEvent = event.Copy(*newEvent, field)
	}

	logp.Debug("filter", "after applying include_fields: %v\n", newEvent)
	event = newEvent
}

func (f *IncludeFields) String() string {
	return "include_fields=" + strings.Join(f.Fields, ", ")
}

/* DropFields methods */
func NewDropFields(fields []string) *DropFields {

	/* remove read only fields */
	for _, readOnly := range ReadOnlyFields {
		for i, field := range fields {
			if readOnly == field {
				fields = append(fields[:i], fields[i+1:]...)
			}
		}
	}
	return &DropFields{Fields: fields}
}

func (f *DropFields) Filter(event *common.MapStr) {

	for _, field := range f.Fields {
		event.Delete(field)

	}
}

func (f *DropFields) String() string {

	return "drop_fields=" + strings.Join(f.Fields, ", ")
}
