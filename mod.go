package keen

type UtilRecord func(string)

var H UtilRecord = func(s string) {
	println(s)
}

func Mute() {
	H = func(s string) {}
}

func Custom(rcd UtilRecord) {
	H = rcd
}
