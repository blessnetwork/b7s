package pbft

type signable interface {
	signableRecord() any
	setSignature(string)
	getSignature() string
}

// Returns the payload that is eligible to be signed. This means basically the PrePrepare struct, excluding the signature field.
func (p *PrePrepare) signableRecord() any {
	cp := *p
	cp.setSignature("")
	return cp
}

func (p *PrePrepare) setSignature(signature string) {
	p.Signature = signature
}

func (p PrePrepare) getSignature() string {
	return p.Signature
}

// Returns the payload that is eligible to be signed. This means basically the Prepare struct, excluding the signature field.
func (p *Prepare) signableRecord() any {
	cp := *p
	cp.setSignature("")
	return cp
}

func (p *Prepare) setSignature(signature string) {
	p.Signature = signature
}

func (p Prepare) getSignature() string {
	return p.Signature
}

// Returns the payload that is eligible to be signed. This means basically the Commit struct, excluding the signature field.
func (c *Commit) signableRecord() any {
	cp := *c
	cp.setSignature("")
	return cp
}

func (c *Commit) setSignature(signature string) {
	c.Signature = signature
}

func (c Commit) getSignature() string {
	return c.Signature
}

// Returns the payload that is eligible to be signed. This means basically the ViewChange struct, excluding the signature field.
func (v *ViewChange) signableRecord() any {
	cp := *v
	cp.setSignature("")
	return cp
}

func (v *ViewChange) setSignature(signature string) {
	v.Signature = signature
}

func (v ViewChange) getSignature() string {
	return v.Signature
}

// Returns the payload that is eligible to be signed. This means basically the NewView struct, excluding the signature field.
func (v *NewView) signableRecord() any {
	cp := *v
	cp.setSignature("")
	return cp
}

func (v *NewView) setSignature(signature string) {
	v.Signature = signature
}

func (v NewView) getSignature() string {
	return v.Signature
}
