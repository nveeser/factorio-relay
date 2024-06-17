package relay

import (
	"fmt"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"strings"
)

func foo() {
	dumpCap()
	capSet := cap.GetProc()
	if err := capSet.SetFlag(cap.Permitted, true, cap.NET_RAW); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Permitted, err)
	}
	if err := capSet.SetProc(); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Permitted, err)
	}
	dumpCap(cap.NET_RAW)
	capSet = cap.GetProc()
	if err := capSet.SetFlag(cap.Effective, true, cap.NET_RAW); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Effective, err)
	}
	if err := capSet.SetProc(); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Effective, err)
	}
	dumpCap(cap.NET_RAW)
	capSet = cap.GetProc()
	if err := capSet.SetFlag(cap.Inheritable, true, cap.NET_RAW); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Inheritable, err)
	}
	if err := capSet.SetProc(); err != nil {
		fmt.Printf("SetFlag(%s): error %s\n", cap.Inheritable, err)
	}
	dumpCap(cap.NET_RAW)

}

func dumpCap(values ...cap.Value) {
	cl := capIterator(values)
	cl.visit(func(value cap.Value) {
		cs, err := capString(value)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fmt.Printf("CAP: %s", cs)
	})
}

type capIterator []cap.Value

func (c capIterator) visit(f func(cap.Value)) {
	if len(c) == 0 {
		for v := cap.Value(0); v < cap.MaxBits(); v++ {
			f(v)
		}
		return
	}
	for _, v := range c {
		f(v)
	}
}

func capString(val cap.Value) (string, error) {
	var vals []string
	for _, flag := range []fmt.Stringer{
		cap.Permitted,
		cap.Effective,
		cap.Inheritable,
		cap.Amb,
		cap.Inh,
		cap.Bound,
	} {
		v, err := flagString(flag, val)
		if err != nil {

			return "", err
		}
		vals = append(vals, v)
	}
	return fmt.Sprintf("[%s] %s\n", strings.Join(vals, ""), val), nil
}

func flagString(flag fmt.Stringer, val cap.Value) (string, error) {
	var present bool
	var err error
	switch x := flag.(type) {
	case cap.Flag:
		procSet := cap.GetProc()
		present, err = procSet.GetFlag(x, val)
		if err != nil {
			return "", fmt.Errorf("error GetFlag(): %w", err)
		}
	case cap.Vector:
		proc := cap.IABGetProc()
		present, err = proc.GetVector(x, val)
		if err != nil {
			return "", fmt.Errorf("error GetVector(): %w", err)
		}
	}
	if present {
		return strings.ToUpper(flag.String()), nil
	} else {
		return strings.ToLower(flag.String()), nil
	}
}
