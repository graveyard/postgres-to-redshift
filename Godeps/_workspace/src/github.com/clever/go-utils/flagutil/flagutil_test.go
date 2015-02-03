package flagutil

import (
	"flag"
	"strings"
	"testing"
)

// ValidateFlags tests.

type nonGetterValue string

func (v *nonGetterValue) String() string {
	return string(*v)
}

func (v *nonGetterValue) Set(s string) error {
	*v = nonGetterValue(s)
	return nil
}

func TestValidateFlagsNoErrorIfAllFlagsSet(t *testing.T) {
	fs := flag.NewFlagSet("TestValidateFlagsNoErrorIfAllFlagsSet", flag.ContinueOnError)
	fs.String("strflag", "", "Example string flag")
	var ngv nonGetterValue
	fs.Var(&ngv, "nongettervalue", "Usage2")
	args := []string{"-strflag=foo", "-nongettervalue=bar"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	if err := ValidateFlags(fs); err != nil {
		t.Fatalf("Error while validating flags. %s", err.Error())
	}
}

func TestValidateFlagsNoErrorIfNonGetterNotSet(t *testing.T) {
	fs := flag.NewFlagSet("TestValidateFlagsNoErrorIfNonGetterNotSet", flag.ContinueOnError)
	fs.String("strflag", "", "Example string flag")
	var ngv nonGetterValue
	fs.Var(&ngv, "nongettervalue", "Usage2")
	args := []string{}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	if err := ValidateFlags(fs); err != nil {
		t.Fatalf("Error while validating flags. %s", err.Error())
	}
}

type getErr string

func (err getErr) Error() string {
	return string(err)
}

type errorGetter string

func (g *errorGetter) String() string {
	return string(*g)
}

func (g *errorGetter) Set(s string) error {
	*g = errorGetter(s)
	return nil
}

func (g *errorGetter) Get() interface{} {
	return getErr("Error")
}

func TestValidateFlagsErrorIfGetError(t *testing.T) {
	fs := flag.NewFlagSet("TestValidateFlagsErrorIfGetError", flag.ContinueOnError)
	var eg1, eg2 errorGetter
	fs.Var(&eg1, "eg1", "Usage1")
	fs.Var(&eg2, "eg2", "Usage2")
	args := []string{"-eg1=foo"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	mferr := ValidateFlags(fs)
	if mferr == nil {
		t.Fatal("Error not returned while validating flags.")
	}
	if len(mferr) != 2 {
		t.Fatalf("Incorrect length of multi flag errors. Expected: 2, Actual: %d", len(mferr))
	}
}

type nonErrorGetter string

func (g *nonErrorGetter) String() string {
	return string(*g)
}

func (g *nonErrorGetter) Set(s string) error {
	*g = nonErrorGetter(s)
	return nil
}

func (g *nonErrorGetter) Get() interface{} {
	return string(*g)
}

func TestValidateFlagsNoErrorIfNoGetError(t *testing.T) {
	fs := flag.NewFlagSet("TestValidateFlagsNoErrorIfNoGetError", flag.ContinueOnError)
	var neg1, neg2 nonErrorGetter
	fs.Var(&neg1, "neg1", "Usage1")
	fs.Var(&neg2, "neg2", "Usage2")
	args := []string{"-neg1=foo"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	if err := ValidateFlags(fs); err != nil {
		t.Fatal("Error returned while validating flags with no get errors. %s", err.Error())
	}
}

// Required String tests.

func TestRequiredStringErrorValidateIfNotSet(t *testing.T) {
	fs := flag.NewFlagSet("TestRequiredStringErrorValidateIfNotSet", flag.ContinueOnError)
	fs.String("strflag", "", "Example string flag")
	RequiredStringFlag("reqstr", "Required String", fs)
	args := []string{"-strflag=foo"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	if err := ValidateFlags(fs); err == nil {
		t.Fatal("Error not returned while validating flags.")
	}
}

func TestRequiredStringParse(t *testing.T) {
	fs := flag.NewFlagSet("TestRequiredStringParse", flag.ContinueOnError)
	rs := RequiredStringFlag("reqstr", "Usage", fs)
	args := []string{"-reqstr=foo"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Error while parsing args %s. %s", strings.Join(args, " "), err.Error())
	}
	if *rs != "foo" {
		t.Fatalf("Required string flag not parsed correctly. Expected: %s. Actual: %s", "foo", *rs)
	}
	if err := ValidateFlags(fs); err != nil {
		t.Fatalf("Error while validating args. %s", err.Error())
	}
}
