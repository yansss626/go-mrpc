package generator

type IDL struct {
	Package string `yaml:"package"`

	Messages map[string]Message `yaml:"messages"`

	Services map[string]Service `yaml:"services"`
}

type Message struct {
	Fields []Field `yaml:"fields"`
}
type Field struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Repeated bool   `yaml:"repeated"`
	Optional bool   `yanml:"optional"`
	JsonName string `yaml:"json_name"`
}
type Service struct {
	Methods map[string]Method `yaml:"methods"`
}

type Method struct {
	Request string `yaml:"request"`

	Response     string `yaml:"response"`
	ServerStream bool   `yaml:"server_stream"`
}
