package lrparser

//// commented: using grammar definition
//func parseLetter(ps *PState) error {
//	ps2 := ps.Copy()
//	ru, err := ps2.ReadRune()
//	if err != nil {
//		return err
//	}
//	if !unicode.IsLetter(ru) {
//		return errors.New("not a letter")
//	}
//	ps.Set(ps2)
//	return nil
//}

//// commented: using grammar definition
//func parseDigit(ps *PState) error {
//	ps2 := ps.Copy()
//	ru, err := ps2.ReadRune()
//	if err != nil {
//		return err
//	}
//	if !unicode.IsDigit(ru) {
//		return errors.New("not a digit")
//	}
//	ps.Set(ps2)
//	return nil
//}

// commented: using this won't recognize "digit" in "digits", which won't allow to parse correctly in some cases
//func parseDigits(ps *PState) error {
//	for i := 0; ; i++ {
//		ps2 := ps.copy()
//		ru, err := ps2.readRune()
//		if err != nil {
//			if i > 0 {
//				return nil
//			}
//			return err
//		}
//		if !unicode.IsDigit(ru) {
//			if i == 0 {
//				return errors.New("not a digit")
//			}
//			return nil
//		}
//		ps.set(ps2)
//	}
//}
