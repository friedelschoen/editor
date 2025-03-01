package goutil

//func FindGoSum(dir string) (string, bool) {
//	return findFileUp(dir, "go.sum")
//}

//func GoModCreateContent(dir string, content string) error {
//	filename := filepath.Join(dir, "go.mod")
//	f, err := os.Create(filename)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	if _, err := fmt.Fprintf(f, content); err != nil {
//		return err
//	}
//	return nil
//}

//func goModReplaceUsingAppend(ctx context.Context, dir, old, new string) error {
//	filename := filepath.Join(dir, "go.mod")
//	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	u := "replace " + old + " => " + new
//	if _, err := f.WriteString("\n" + u + "\n"); err != nil {
//		return err
//	}
//	return nil
//}
