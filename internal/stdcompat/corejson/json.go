// SPDX-License-Identifier: EUPL-1.2

package stdcompat

import core "dappco.re/go"

func Marshal(v any) (
	[]byte,
	error,
) {
	r := core.JSONMarshal(v)
	if !r.OK {
		err, _ := r.Value.(error)
		return nil, err
	}
	data, _ := r.Value.([]byte)
	return data, nil
}

func Unmarshal(data []byte, target any) (
	_ error,
) {
	r := core.JSONUnmarshal(data, target)
	if r.OK {
		return nil
	}
	err, _ := r.Value.(error)
	return err
}

type Encoder struct{ w core.Writer }
type Decoder struct{ r core.Reader }

func NewEncoder(w core.Writer) *Encoder { return &Encoder{w: w} }
func NewDecoder(r core.Reader) *Decoder { return &Decoder{r: r} }

func (e *Encoder) Encode(v any) (
	_ error,
) {
	_, err := e.w.Write([]byte(core.JSONMarshalString(v) + "\n"))
	return err
}

func (d *Decoder) Decode(v any) (
	_ error,
) {
	r := core.ReadAll(d.r)
	if !r.OK {
		err, _ := r.Value.(error)
		return err
	}
	text, _ := r.Value.(string)
	return Unmarshal([]byte(text), v)
}
