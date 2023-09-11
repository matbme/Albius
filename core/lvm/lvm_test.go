package lvm

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	albius "github.com/vanilla-os/albius/core"
)

var (
	lvm     Lvm
	device  string
	lvmpart string
)

func TestMain(m *testing.M) {
	// Create LVM wrapper instance
	lvm = NewLvm()

	// Setup testing device
	// Create dummy image
	cmd := exec.Command("dd", "if=/dev/zero", "of=test.img", "count=102400")
	if err := cmd.Run(); err != nil {
		panic("error while creating testing device image: " + err.Error())
	}
	// Mount dummy image as loop device
	cmd = exec.Command("losetup", "--find", "--show", "test.img")
	cmd.Stderr = os.Stderr
	ret, err := cmd.Output()
	if err != nil {
		panic("error while mounting loop device: " + err.Error())
	}
	device = string(ret)
	device = device[:len(device)-1]
	//Create device label and add some partitions
	albiusDevice, err := albius.LocateDisk(device)
	if err != nil {
		panic("error finding loop device: " + err.Error())
	}
	err = albiusDevice.LabelDisk(albius.GPT)
	if err != nil {
		panic("error adding label to loop device: " + err.Error())
	}
	_, err = albiusDevice.NewPartition("", albius.EXT4, 1, 25)
	if err != nil {
		panic("error creating partition A in loop device: " + err.Error())
	}
	_, err = albiusDevice.NewPartition("", albius.EXT4, 26, -1)
	if err != nil {
		panic("error creating partition B in loop device: " + err.Error())
	}
	lvmpart = device + "p"

	// Run tests
	status := m.Run()

	// Remove testing device
	cmd = exec.Command("losetup", "-d", device)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("error while detaching testing device: " + err.Error())
	}
	err = os.Remove("test.img")
	if err != nil {
		panic("error while removing testing device image: " + err.Error())
	}

	// Cleanup
	lvm.Dispose()
	os.Exit(status)
}

func TestPvcreate(t *testing.T) {
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvs(t *testing.T) {
	pvs, err := lvm.Pvs()
	fmt.Printf(" -> Returned: %v\n", pvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvResize(t *testing.T) {
	pvs, err := lvm.Pvs()
	if err != nil {
		t.Fatal(err)
	}
	err = lvm.Pvresize(&pvs[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvShrink(t *testing.T) {
	pvs, err := lvm.Pvs()
	if err != nil {
		t.Fatal(err)
	}
	err = lvm.Pvresize(&pvs[0], 10.0)
	if err != nil {
		t.Fatal(err)
	}

	pvs, err = lvm.Pvs()
	fmt.Printf(" -> New size: %v\n", pvs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvRemoveStr(t *testing.T) {
	err := lvm.Pvremove(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPvRemoveStruct(t *testing.T) {
	// Recreate PV removed by previous test
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}

	pvs, err := lvm.Pvs()
	if err != nil {
		t.Fatal(err)
	}

	err = lvm.Pvremove(pvs[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgCreate(t *testing.T) {
	// Create two testing PVs
	err := lvm.Pvcreate(lvmpart + "1")
	if err != nil {
		t.Fatal(err)
	}
	err = lvm.Pvcreate(lvmpart + "2")
	if err != nil {
		t.Fatal(err)
	}

	// Pass one PV as struct and another as string
	pvs, err := lvm.Pvs()
	if err != nil {
		t.Fatal(err)
	}

	err = lvm.Vgcreate("MyTestingVG", pvs[0], lvmpart+"2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestVgs(t *testing.T) {
	vgs, err := lvm.Vgs()
	fmt.Printf(" -> Returned: %v\n", vgs)
	if err != nil {
		t.Fatal(err)
	}
}
