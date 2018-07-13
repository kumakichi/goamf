package amf

const (
	TYPE_NUMBER      = 0x00
	TYPE_BOOL        = 0x01
	TYPE_STRING      = 0x02
	TYPE_OBJECT      = 0x03
	TYPE_MOVIECLIP   = 0x04
	TYPE_NULL        = 0x05
	TYPE_UNDEFINED   = 0x06
	TYPE_REFERENCE   = 0x07
	TYPE_MIXEDARRAY  = 0x08
	TYPE_OBJECTTERM  = 0x09
	TYPE_ARRAY       = 0x0A
	TYPE_DATE        = 0x0B
	TYPE_LONGSTRING  = 0x0C
	TYPE_UNSUPPORTED = 0x0D
	TYPE_RECORDSET   = 0x0E
	TYPE_XML         = 0x0F
	TYPE_TYPEDOBJECT = 0x10
	TYPE_AMF3        = 0x11
)

func (cxt *Decoder) ReadVal() interface{} {
	marker := cxt.ReadByte()

	switch marker {
	case TYPE_STRING:
		_, v := cxt.ReadString()
		return v
	case TYPE_AMF3:
		cxt.IsAMF3 = true
		return cxt.ReadValueAmf3()
		// .... 太多，暂时用不上，不写了
	}
	return nil
}
