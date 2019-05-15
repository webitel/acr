package call

import "regexp"

type CallRegExp map[string][]string

func (r CallRegExp) Get(position string, idx int) string {
	if v, ok := r[position]; ok {
		if len(v) > idx {
			return v[idx]
		}
	}

	return ""
}

func (call *Call) AddRegExp(data []string) {
	call.regExp["reg_"+string(len(call.regExp))] = data
}

func setupNumber(reg, dest string) map[string][]string {
	storage := make(map[string][]string)
	if reg != "" {
		re, err := regexp.Compile(reg)
		if err == nil {
			d := re.FindStringSubmatch(dest)
			storage["0"] = d
		}
	} else {
		storage["0"] = []string{}
	}
	return storage
}
