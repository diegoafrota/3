package script

import (
	"log"
	"math"
)

// Loads standard functions into the world.
func (w *World) LoadStdlib() {

	// literals
	w.declare("true", boolLit(true))
	w.declare("false", boolLit(false))

	// io
	w.Func("print", myprint)

	// math
	w.Func("Square", square)
	w.Func("Abs", math.Abs)
	w.Func("Acos", math.Acos)
	w.Func("Acosh", math.Acosh)
	w.Func("Asin", math.Asin)
	w.Func("Asinh", math.Asinh)
	w.Func("Atan", math.Atan)
	w.Func("Atanh", math.Atanh)
	w.Func("Cbrt", math.Cbrt)
	w.Func("Ceil", math.Ceil)
	w.Func("Cos", math.Cos)
	w.Func("Cosh", math.Cosh)
	w.Func("Erf", math.Erf)
	w.Func("Erfc", math.Erfc)
	w.Func("Exp", math.Exp)
	w.Func("Exp2", math.Exp2)
	w.Func("Expm1", math.Expm1)
	w.Func("Floor", math.Floor)
	w.Func("Gamma", math.Gamma)
	w.Func("J0", math.J0)
	w.Func("J1", math.J1)
	w.Func("Log", math.Log)
	w.Func("Log10", math.Log10)
	w.Func("Log1p", math.Log1p)
	w.Func("Log2", math.Log2)
	w.Func("Logb", math.Logb)
	w.Func("Sin", math.Sin)
	w.Func("Sinh", math.Sinh)
	w.Func("Sqrt", math.Sqrt)
	w.Func("Tan", math.Tan)
	w.Func("Tanh", math.Tanh)
	w.Func("Trunc", math.Trunc)
	w.Func("Y0", math.Y0)
	w.Func("Y1", math.Y1)
	w.Func("Ilogb", math.Ilogb)
	w.Func("Pow10", math.Pow10)
	w.Func("Atan2", math.Atan2)
	w.Func("Hypot", math.Hypot)
	w.Func("Remainder", math.Remainder)
	w.Func("Max", math.Max)
	w.Func("Min", math.Min)
	w.Func("Mod", math.Mod)
	w.Func("Pow", math.Pow)
	w.Func("Yn", math.Yn)
	w.Func("Jn", math.Jn)
	w.Func("Ldexp", math.Ldexp)
	w.Func("IsInf", math.IsInf)
	w.Func("IsNaN", math.IsNaN)
	w.declare("Pi", floatLit(math.Pi))
	w.declare("Inf", floatLit(math.Inf(1)))
}

func myprint(msg ...interface{}) {
	log.Println(msg...)
}

func square(x float64) float64 {
	return x * x
}
