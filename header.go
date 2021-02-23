package esl

type Header struct {
	contentLength int

	args Args
}

func (h *Header) Set(key, value string) {
	h.args.Set(key, value)
}

func (h *Header) Add(key, value string) {
	h.args.Add(key, value)
}

func (h *Header) AddBytes(key, value []byte) {
	h.args.AddBytes(key, value)
}

func (h *Header) GetInt(key string) (int, error) {
	return h.args.GetInt([]byte(key))
}

func (h *Header) Get(key string) string {
	return string(h.args.GetBytes([]byte(key)))
}

func (h *Header) ContentLength() (int, error) {
	if h.contentLength != -1 {
		return h.contentLength, nil
	}
	l, err := h.GetInt("Content-Length")
	h.contentLength = l
	return l, err
}

func (h *Header) reset() {
	h.contentLength = -1
	h.args.reset()
}
