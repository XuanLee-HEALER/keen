package keen

type UtilRecord func(string)

var H UtilRecord = func(s string) {
	println(s)
}

func Mute() {
	H = nil
}

func Custom(rcd UtilRecord) {
	H = rcd
}
