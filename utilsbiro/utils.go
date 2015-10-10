package utilsbiro

/**
 * Do a obvious conversion for a lazy language!
 */
func ToInterfaceArr(strArr []string) []interface{} {

	interfaceParam := make([]interface{}, len(strArr))

	for i, p := range strArr {
		interfaceParam[i] = p
	}

	return interfaceParam

}
