package esl

import "fmt"

// Key identifier.
type Key int

func ( id Key ) Error( ) string {
	return fmt.Sprintf( "!Key(%+d)", id )
}

// Errorf ...
func ( id Key ) Errorf( message string, arg ...interface{ } ) error {
	return fmt.Errorf( message, arg...)
}