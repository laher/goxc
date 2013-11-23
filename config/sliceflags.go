package config

import (
	"fmt"
)

// Define a type named "intslice" as a slice of ints
type Strslice []string

// Now, for our new type, implement the two methods of
// the flag.Value interface...
// The first method is String() string
func (i *Strslice) String() string {
	return fmt.Sprintf("%v", *i)
}

// The second method is Set(value string) error
func (i *Strslice) Set(value string) error {
	fmt.Printf("Adding %s\n", value)
	*i = append(*i, value)
	return nil
}
