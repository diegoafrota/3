package engine

import (
	"github.com/barnex/cuda5/cu"
	"github.com/mumax/3/cuda"
	"github.com/mumax/3/data"
	"github.com/mumax/3/util"
	"math"
	"unsafe"
)

var (
	Aex        ScalarParam // Exchange stiffness
	Dex        VectorParam // DMI strength
	B_exch     adder       // exchange field (T) output handle
	lex2       exchParam   // inter-cell exchange length squared * 1e18
	E_exch     *GetScalar
	Edens_exch adder
)

func init() {
	Aex.init("Aex", "J/m", "Exchange stiffness", []derived{&lex2})
	Dex.init("Dex", "J/m2", "Dzyaloshinskii-Moriya strength")
	B_exch.init(VECTOR, "B_exch", "T", "Exchange field", AddExchangeField)
	E_exch = NewGetScalar("E_exch", "J", "Exchange energy (normal+DM)", GetExchangeEnergy)
	Edens_exch.init(SCALAR, "Edens_exch", "J/m3", "Exchange energy density (normal+DM)", addEdens(&B_exch, -0.5))
	registerEnergy(GetExchangeEnergy, Edens_exch.AddTo)
	DeclFunc("SetExLen", OverrideExchangeLength, "Sets inter-material exchange length between two regions.")
	lex2.init()
}

// Adds the current exchange field to dst
func AddExchangeField(dst *data.Slice) {
	if Dex.isZero() {
		cuda.AddExchange(dst, M.Buffer(), lex2.Gpu(), regions.Gpu(), M.Mesh())
	} else {
		// DMI only implemented for uniform parameters
		// interaction not clear with space-dependent parameters
		util.AssertMsg(Mesh().Size()[Z] == 1,
			"DMI: mesh must be 2D")
		util.AssertMsg(Msat.IsUniform() && Aex.IsUniform() && Dex.IsUniform(),
			"DMI: Msat, Aex, Dex must be uniform")
		msat := Msat.GetRegion(0)
		D := Dex.GetRegion(0)
		A := Aex.GetRegion(0) / msat
		cuda.AddDMI(dst, M.Buffer(), float32(D[X]/msat), float32(D[Y]/msat), float32(D[Z]/msat), float32(A), M.Mesh()) // dmi+exchange
	}
}

// Returns the current exchange energy in Joules.
// Note: the energy is defined up to an arbitrary constant,
// ground state energy is not necessarily zero or comparable
// to other simulation programs.
func GetExchangeEnergy() float64 {
	return -0.5 * cellVolume() * dot(&M_full, &B_exch)
}

// Defines the exchange coupling between different regions by specifying the
// exchange length of the interaction between them.
// 	lex := sqrt(2*Aex / Msat)
// In case of different materials it is not always clear what the exchange
// between them should be, especially if they have different Msat. By specifying
// the exchange length, it is up to the user to decide which Msat to use.
// A negative length may be specified to obtain antiferromagnetic coupling.
func OverrideExchangeLength(region1, region2 int, exlen float64) {
	l2 := sign(exlen) * (exlen * exlen) * 1e18
	lex2.override[symmidx(region1, region2)] = float32(l2)
	lex2.gpu_ok = false
}

// stores interregion exchange stiffness
type exchParam struct {
	lut, override  [NREGION * (NREGION + 1) / 2]float32 // cpu lookup-table
	gpu            cuda.SymmLUT                         // gpu copy of lut, lazily transferred when needed
	gpu_ok, cpu_ok bool                                 // gpu cache up-to date with lut source
}

func (p *exchParam) invalidate() { p.cpu_ok = false }

const no_override = math.MaxFloat32

func (p *exchParam) init() {
	for i := range p.override {
		p.override[i] = no_override // dummy value means no override
	}
}

// Get a GPU mirror of the look-up table.
// Copies to GPU first only if needed.
func (p *exchParam) Gpu() cuda.SymmLUT {
	p.update()
	if !p.gpu_ok {
		p.upload()
	}
	return p.gpu
}

func (p *exchParam) update() {
	if !p.cpu_ok {
		msat := Msat.cpuLUT()
		aex := Aex.cpuLUT()

		// todo: conditional
		for i := 0; i < NREGION; i++ {
			lexi := 2e18 * safediv(aex[0][i], msat[0][i])
			for j := 0; j <= i; j++ {
				I := symmidx(i, j)
				if p.override[I] == no_override {
					lexj := 2e18 * safediv(aex[0][j], msat[0][j])
					p.lut[I] = 2 / (1/lexi + 1/lexj)
				} else {
					p.lut[I] = p.override[I]
				}
			}
		}
		p.gpu_ok = false
		p.cpu_ok = true
	}
}

func (p *exchParam) upload() {
	// alloc if  needed
	if p.gpu == nil {
		p.gpu = cuda.SymmLUT(cuda.MemAlloc(int64(len(p.lut)) * cu.SIZEOF_FLOAT32))
	}
	// TODO: sync?
	cu.MemcpyHtoD(cu.DevicePtr(p.gpu), unsafe.Pointer(&p.lut[0]), cu.SIZEOF_FLOAT32*int64(len(p.lut)))
	p.gpu_ok = true
}

func (p *exchParam) SetInterRegion(r1, r2 int, val float64) {
	v := float32(val)
	p.lut[symmidx(r1, r2)] = v

	if r1 == r2 {
		r := r1
		for i := 0; i < NREGION; i++ {
			if p.lut[symmidx(i, i)] == v {
				p.lut[symmidx(r, i)] = v
			} else {
				p.lut[symmidx(r, i)] = 0 // TODO: harmnoic avg !!!
			}
		}
	}

	p.gpu_ok = false
}

// Index in symmetric matrix where only one half is stored.
// (!) Code duplicated in exchange.cu
func symmidx(i, j int) int {
	if j <= i {
		return i*(i+1)/2 + j
	} else {
		return j*(j+1)/2 + i
	}
}

//func (p *exchParam) String() string {
//	str := ""
//	for j := 0; j < NREGION; j++ {
//		for i := 0; i <= j; i++ {
//			str += fmt.Sprint(p.lut[symmidx(i, j)], "\t")
//		}
//		str += "\n"
//	}
//	return str
//}
