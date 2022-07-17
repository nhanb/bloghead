package main

func check(e error) {
	// Gonna default to panicking at every error until I finally bother to
	// learn The Right Way (tm) to juggle them in Go.
	if e != nil {
		panic(e)
	}
}
