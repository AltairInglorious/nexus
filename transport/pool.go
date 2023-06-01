package transport

func (t *Transport) getErrorFromPool(status int, errMsg string) *NATSError {
	err := t.errPool.Get().(*NATSError)
	err.Status = status
	err.Error = errMsg
	return err
}

func (t *Transport) returnErrorToPool(err *NATSError) {
	err.Status = 0
	err.Error = ""
	t.errPool.Put(err)
}

func (t *Transport) getOkFromPool(status int, body any) *NATSOk {
	o := t.okPool.Get().(*NATSOk)
	o.Status = status
	o.Body = body
	return o
}

func (t *Transport) returnOkToPool(o *NATSOk) {
	o.Status = 0
	o.Body = nil
	t.errPool.Put(o)
}
