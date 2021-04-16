package truetype

var _ VariableFont = (*FontMetrics)(nil)

func (f *FontMetrics) Variations() TableFvar { return f.fvar }

// VariableFont is implemented by formats with variable font
// support.
// TODO: polish
type VariableFont interface {
	Variations() TableFvar
}

type VarAxis struct {
	Tag     Tag     // Tag identifying the design variation for the axis.
	Minimum float32 // mininum value on the variation axis that the font covers
	Default float32 // default position on the axis
	Maximum float32 // maximum value on the variation axis that the font covers
	flags   uint16  // Axis qualifiers — see details below.
	strid   NameID  // name entry in the font's ‘name’ table
}

type VarInstance struct {
	Coords    []float32 // length: number of axis
	Subfamily NameID

	PSStringID NameID
}

type TableFvar struct {
	Axis      []VarAxis
	Instances []VarInstance // contains the default instance
}

// IsDefaultInstance returns `true` is `instance` has the same
// coordinates as the default instance.
func (t TableFvar) IsDefaultInstance(it VarInstance) bool {
	for i, c := range it.Coords {
		if c != t.Axis[i].Default {
			return false
		}
	}
	return true
}

// add the default instance if it not already explicitely present
func (t *TableFvar) checkDefaultInstance(names TableName) {
	for _, instance := range t.Instances {
		if t.IsDefaultInstance(instance) {
			return
		}
	}

	// add the default instance
	// choose the subfamily entry
	subFamily := NamePreferredSubfamily
	if v1, v2 := names.getEntry(subFamily); v1 == nil && v2 == nil {
		subFamily = NameFontSubfamily
	}
	defaultInstance := VarInstance{
		Coords:     make([]float32, len(t.Axis)),
		Subfamily:  subFamily,
		PSStringID: NamePostscript,
	}
	for i, axe := range t.Axis {
		defaultInstance.Coords[i] = axe.Default
	}
	t.Instances = append(t.Instances, defaultInstance)
}

// FindAxis return the axis for the given tag and its index, or -1 if not found.
func (t *TableFvar) FindAxis(tag Tag) (VarAxis, int) {
	for i, axis := range t.Axis {
		if axis.Tag == tag {
			return axis, i
		}
	}
	return VarAxis{}, -1
}
