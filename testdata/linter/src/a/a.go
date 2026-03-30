package main

func Bad() {
	panic("test") // want `panic not allowed`
}
