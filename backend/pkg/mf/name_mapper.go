package mf


var name_mapper = map[string]string{
	
}
func GetMFCode(name string) string {
	return name_mapper[name]
}
