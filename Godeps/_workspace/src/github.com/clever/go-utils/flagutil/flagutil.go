// Package flagutil contains flag utilities like implementations of flag.Value and flag.Getter
// interfaces. Also contains methods like ValidateFlags.
package flagutil

import (
	"flag"
	"fmt"
	"strings"
)

// FlagError is the error returned while validating a single flag.
type FlagError struct {
	Flag *flag.Flag
	Err  error
}

func (f *FlagError) Error() string {
	return fmt.Sprintf("Flag: %s, Error: %s, Usage: %s", f.Flag.Name, f.Err.Error(), f.Flag.Usage)
}

// MultiFlagError is a slice of errors returned while validating flags.
type MultiFlagError []*FlagError

func (m MultiFlagError) Error() string {
	errorstr := []string{"Errors found while validating flags"}
	for _, err := range m {
		errorstr = append(errorstr, err.Error())
	}
	return strings.Join(errorstr, "\n")
}

// ValidateFlags validates flags and logs a fatal message with all errors if found. Must be called
// after flag.Parse(). This is done by calling Get() on all flag values which implement the
// flag.Getter interface and collecting errors if returned.
func ValidateFlags(fs *flag.FlagSet) MultiFlagError {
	if fs == nil {
		fs = flag.CommandLine
	}
	var flagErrors []*FlagError
	checkGetError := func(f *flag.Flag) {
		if fv, ok := f.Value.(flag.Getter); ok {
			if err, ok := fv.Get().(error); ok && err != nil {
				flagErrors = append(flagErrors, &FlagError{f, err})
			}
		}
	}
	fs.VisitAll(checkGetError)
	if len(flagErrors) > 0 {
		return MultiFlagError(flagErrors)
	}
	return nil
}

type flagValueError string

func (err flagValueError) Error() string {
	return string(err)
}

type requiredString struct {
	str *string
	set bool
}

// RequiredStringFlag adds a required string flag with the specified name. Note that ValidateFlags
// must be called to ensure that it is set. Returns a pointer to the string value set.
// Example usage:
// var reqstr = flagutil.RequiredStringFlag(name, usage, nil)
// func main() {
// 	flag.Parse()
// 	if err := flagutil.ValidateFlags(nil); err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println(*reqstr)
// }
func RequiredStringFlag(name string, usage string, fs *flag.FlagSet) *string {
	if fs == nil {
		fs = flag.CommandLine
	}
	rs := requiredString{new(string), false}
	fs.Var(&rs, name, usage)
	return rs.str
}

func (rs *requiredString) Set(s string) error {
	*(rs.str), rs.set = s, true
	return nil
}

func (rs *requiredString) Get() interface{} {
	if !rs.set {
		return flagValueError("FlagValue of type requiredString not set")
	}
	return *(rs.str)
}

func (rs *requiredString) String() string {
	return fmt.Sprintf("%s", *(rs.str))
}
