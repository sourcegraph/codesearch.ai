func (enc *StreamEncoder) Encode(val interface{}) (err error) {
	out := newBytes()

	/* encode into the buffer */
	err = EncodeInto(&out, val, enc.Opts)
	if err != nil {
		goto free_bytes
	}

	if enc.indent != "" || enc.prefix != "" {
		/* indent the JSON */
		buf := newBuffer()
		err = json.Indent(buf, out, enc.prefix, enc.indent)
		if err != nil {
			freeBuffer(buf)
			goto free_bytes
		}

		// according to standard library, terminate each value with a newline...
		buf.WriteByte('\n')

		/* copy into io.Writer */
		_, err = io.Copy(enc.w, buf)
		if err != nil {
			freeBuffer(buf)
			goto free_bytes
		}

	} else {
		/* copy into io.Writer */
		var n int
		for len(out) > 0 {
			n, err = enc.w.Write(out)
			out = out[n:]
			if err != nil {
				goto free_bytes
			}
		}

		// according to standard library, terminate each value with a newline...
		enc.w.Write([]byte{'\n'})
	}

free_bytes:
	freeBytes(out)
	return err
}