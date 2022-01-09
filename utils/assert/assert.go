package assert

func assert(exp bool) {
	if !exp {
		panic("")
	}
}