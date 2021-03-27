package fuzz

var ff = [...]interface{}{
	FuzzPerson,
	FuzzPerson_Male,
	FuzzPerson_Female,
	FuzzPerson_Other,
	FuzzBooks,
}

func FuzzFuncs() []interface{} {
	return ff[:]
}
