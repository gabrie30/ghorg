package markdown

var defaultHeadingShades = []shadeFmt{
	GreenBold,
	GreenBold,
	HiGreen,
	Green,
}

var defaultQuoteShades = []shadeFmt{
	GreenBold,
	GreenBold,
	HiGreen,
	Green,
}

type shadeFmt func(a ...interface{}) string

type levelShadeFmt func(level int) shadeFmt

// Return a function giving the color function corresponding to the level.
// Beware, level start counting from 1.
func shade(shades []shadeFmt) levelShadeFmt {
	return func(level int) shadeFmt {
		if level < 1 {
			level = 1
		}
		if level > len(shades) {
			level = len(shades)
		}
		return shades[level-1]
	}
}
