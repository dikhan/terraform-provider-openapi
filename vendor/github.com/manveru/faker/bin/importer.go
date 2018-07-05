package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "(en|de|nl|...)")
		os.Exit(1)
	}

	final := make(map[string]map[string][]string)

	for _, lang := range os.Args[1:] {
		fmt.Fprintln(os.Stderr, lang)
		f, err := os.Open("/home/manveru/.gem/ruby/1.9.1/gems/faker-1.0.1/lib/locales/" + lang + ".yml")
		if err != nil {
			panic(err)
		}

		in, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}

		final[lang] = make(map[string][]string)

		out := make(map[string]map[string]map[string]map[string]interface{})
		err = yaml.Unmarshal(in, out)
		if err != nil {
			panic(err)
		}

		for _, v1 := range out { // language
			for _, v2 := range v1 { // faker
				for k3, v3 := range v2 {
					for k4, v4 := range v3 {
						key := k3 + "." + k4
						switch v4.(type) {
						case []interface{}:
							for k5, v5 := range v4.([]interface{}) {
								switch v5.(type) {
								case string:
									final[lang][key] = append(final[lang][key], v5.(string))
								case []interface{}:
									arr := make([]string, 0)
									for _, v6 := range v5.([]interface{}) {
										arr = append(arr, v6.(string))
									}
									key5 := fmt.Sprintf("%s.%d", key, k5)
									sort.Strings(arr)
									final[lang][key5] = arr
								default:
									panic(fmt.Sprintf("%#v", v5))
								}
							}
						case map[interface{}]interface{}:
							for k5, v5 := range v4.(map[interface{}]interface{}) {
								switch v5.(type) {
								case []interface{}:
									arr := make([]string, 0)
									for _, v6 := range v5.([]interface{}) {
										arr = append(arr, v6.(string))
									}
									var key5 string
									switch k5.(type) {
									case string:
										key5 = fmt.Sprintf("%s.%s", key, k5)
									default:
										panic(fmt.Sprintf("%#v", k5))
									}
									sort.Strings(arr)
									final[lang][key5] = arr
								default:
									panic("")
								}
							}
						default:
							panic("")
						}
					}
				}
			}
		}
	}

	fmt.Println("package faker\n")
	fmt.Println("var Dict = map[string]map[string][]string{")
	for language, maps := range final {
		fmt.Printf("\t%q: map[string][]string{\n", language)
		for key, values := range maps {
			fmt.Printf("\t\t%q: []string{\n", key)
			for _, value := range values {
				fmt.Printf("\t\t\t%q,\n", value)
			}
			fmt.Println("\t\t},")
		}
		fmt.Println("\t},")
	}
	fmt.Println("}")
}
