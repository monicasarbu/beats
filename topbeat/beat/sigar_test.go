package beat

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSystemLoad(t *testing.T) {

	if runtime.GOOS == "windows" {
		return //no load data on windows
	}

	load, err := GetSystemLoad()

	assert.NotNil(t, load)
	assert.Nil(t, err)

	assert.True(t, (load.Load1 > 0))
	assert.True(t, (load.Load5 > 0))
	assert.True(t, (load.Load15 > 0))
}

func TestGetCpuTimes(t *testing.T) {

	cpu_stat, err := GetCpuTimes()

	assert.NotNil(t, cpu_stat)
	assert.Nil(t, err)

	assert.True(t, (cpu_stat.Cpu.User > 0))
	assert.True(t, (cpu_stat.Cpu.Sys > 0))

}

func TestGetMemory(t *testing.T) {
	mem, err := GetMemory()

	assert.NotNil(t, mem)
	assert.Nil(t, err)

	assert.True(t, (mem.Mem.Total > 0))
	assert.True(t, (mem.Mem.Used > 0))
	assert.True(t, (mem.Mem.Free >= 0))
	assert.True(t, (mem.Mem.ActualFree >= 0))
	assert.True(t, (mem.Mem.ActualUsed > 0))
}

func TestGetSwap(t *testing.T) {

	if runtime.GOOS == "windows" {
		return //no load data on windows
	}

	swap, err := GetSwap()

	assert.NotNil(t, swap)
	assert.Nil(t, err)

	assert.True(t, (swap.Swap.Total >= 0))
	assert.True(t, (swap.Swap.Used >= 0))
	assert.True(t, (swap.Swap.Free >= 0))
}

func TestPids(t *testing.T) {
	pids, err := Pids()

	assert.NotNil(t, pids)
	assert.Nil(t, err)

	// Assuming at least 2 processes are running
	assert.True(t, (len(pids) > 1))
}

func TestGetProcess(t *testing.T) {
	pids, err := Pids()

	assert.Nil(t, err)

	for _, pid := range pids {

		process, err := GetProcess(pid)

		if err != nil {
			continue
		}
		assert.NotNil(t, process)

		assert.True(t, (process.Pid > 0))
		assert.True(t, (process.Ppid >= 0))
		assert.True(t, (len(process.Name) > 0))
		assert.NotEqual(t, "unknown", process.State)

		// Memory Checks
		assert.True(t, (process.Mem.Size >= 0))
		assert.True(t, (process.Mem.Resident >= 0))
		assert.True(t, (process.Mem.Share >= 0))

		// CPU Checks
		assert.True(t, (process.Cpu.StartTime > 0))
		assert.True(t, (process.Cpu.Total >= 0))
		assert.True(t, (process.Cpu.User >= 0))
		assert.True(t, (process.Cpu.Sys >= 0))

		assert.True(t, (process.ctime.Unix() <= time.Now().Unix()))

		// it's enough to get valid data for a single process
		break
	}
}

func TestProcState(t *testing.T) {

	assert.Equal(t, getProcState('R'), "running")
	assert.Equal(t, getProcState('S'), "sleeping")
	assert.Equal(t, getProcState('s'), "unknown")
	assert.Equal(t, getProcState('D'), "idle")
	assert.Equal(t, getProcState('T'), "stopped")
	assert.Equal(t, getProcState('Z'), "zombie")
}

func TestFileSystemList(t *testing.T) {

	if runtime.GOOS == "darwin" && os.Getenv("TRAVIS") == "true" {
		t.Skip("FileSystem test fails on Travis/OSX with i/o error")
	}

	fss, err := GetFileSystemList()

	assert.Nil(t, err)
	assert.True(t, (len(fss) > 0))

	for _, fs := range fss {

		stat, err := GetFileSystemStat(fs)
		assert.NoError(t, err)

		assert.True(t, (stat.Stat.Total >= 0))
		assert.True(t, (stat.Stat.Free >= 0))
		assert.True(t, (stat.Stat.Avail >= 0))
		assert.True(t, (stat.Stat.Used >= 0))
	}
}
