package pargs

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
)

type T interface{}

type Args struct {
	Options          []string
	ParamType        T
	IsFlag           bool
	ExpectParamCount T
	DefaultValue     T
	Help             string
}

type ParamList struct {
	ProgramName string
	List        []*Args
}

func (a *ParamList) StrLen(str string) (length int) {
	for _, v := range str {
		if v > 255 {
			length += 2
		} else {
			length++
		}
	}
	return
}

// 收集所有的参数,然后输出参数说明
func (a *ParamList) CollectArgsInfo(programName string, version string) {
	a.ProgramName = programName
	for _, v := range os.Args {
		if v == "-h" || v == "--help" {
			fmt.Println("ArgsParse Help", version)
			a.Help()
			os.Exit(0)
		}
	}
}

func (p *ParamList) Help() {
	// 默认值的最大长度
	const MAX_VALUE = 20
	// 备注的最大长度
	const MAX_REAMRK = 50

	fmt.Printf("Usage: %s [flag]|[option <args>] ...\n", p.ProgramName)
	// fmt.Printf("-h/--help\t\t%s\n", "帮助")
	maxWidths := make([]int, 6)
	titles := []string{"Options", "ParamType", "IsFlag", "Count", "Default", "Remark"}
	for _, v := range p.List {
		widths := []int{
			len(strings.Join(v.Options, "/")),
			len(reflect.TypeOf(v.ParamType).String()),
			len(fmt.Sprintf("%v", v.IsFlag)),
			len(fmt.Sprintf("%v", v.ExpectParamCount)),
			p.StrLen(fmt.Sprintf("%v", v.DefaultValue)),
			p.StrLen(v.Help),
		}
		if widths[4] > MAX_VALUE {
			widths[4] = MAX_VALUE
		}
		if widths[5] > MAX_REAMRK {
			widths[5] = MAX_REAMRK
		}
		for i, width := range widths {
			if width > maxWidths[i] {
				maxWidths[i] = width
			}
		}
	}
	w := maxWidths
	t := titles
	w1 := make([]int, len(t))
	for i := range maxWidths {
		if maxWidths[i] < len(titles[i]) {
			maxWidths[i] = len(titles[i]) + 2
		}
	}
	for i := range w1 {
		w1[i] = len(titles[i])
	}
	r := func(c int) string {
		return strings.Repeat("-", c)
	}
	_fmtheader := "    \033[35m%-*s  %-*s %-*s %-*s %-*s  %-*s\033[0m\n"
	_fmtstr := "    \033[31m%-*s\033[0m  %-*s %-*s %-*s %-*s  %-*s\n"
	// _fmtheader := "    \033[35m%-15s  %10s %10s %8s %10s   %-15s\033[0m\n"
	// _fmtstr := "    \033[31m%-15s\033[0m  %10s %10s %8s %10s   %-15s\n"
	fmt.Println()
	fmt.Printf(_fmtheader, w[0], t[0], w[1], t[1], w[2], t[2], w[3], t[3], w[4], t[4], w[5], t[5])
	// fmt.Printf(_fmtheader, w[0], r(w1[0]), w[1], r(w1[1]), w[2], r(w1[2]), w[3], r(w1[3]), w[4], r(w1[4]), w[5], r(w[5]))
	fmt.Printf(_fmtheader, w[0], r(w[0]), w[1], r(w[1]), w[2], r(w[2]), w[3], r(w[3]), w[4], r(w[4]), w[5], r(w[5]))
	for _, v := range p.List {
		options := strings.Join(v.Options, "/")
		paramType := reflect.TypeOf(v.ParamType).String()
		flag := "No"
		defaultValue := fmt.Sprintf("%v", v.DefaultValue)
		paramCount := fmt.Sprintf("%v", v.ExpectParamCount)
		if v.IsFlag {
			flag = "Yes"
		}
		if defaultValue == "" {
			defaultValue = "\"\""
		} else if len(defaultValue) > MAX_VALUE {
			defaultValue = defaultValue[:MAX_VALUE]
		}
		if v.Help == "" {
			v.Help = "\"\""
		} else if len(v.Help) > MAX_REAMRK {
			v.Help = v.Help[:MAX_REAMRK]
		}
		fmt.Printf(_fmtstr, w[0], options, w[1], paramType, w[2], flag, w[3], paramCount, w[4], defaultValue, w[5], v.Help)
		// fmt.Printf(_fmtstr, options, paramType, flag, paramCount, defaultValue, v.Help)
	}
	fmt.Println()
}

func (p *ParamList) Flag(options []string, flag *bool, defaultValue bool, help string) {
	*flag = defaultValue
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *flag,
		IsFlag:           true,
		ExpectParamCount: 0,
		DefaultValue:     defaultValue,
		Help:             help,
	})

	osArgs := os.Args
	for _, v := range osArgs {
		for _, o := range options {
			if v == o {
				*flag = true
				break
			}
		}
	}
}

func (p *ParamList) Int(options []string, param *int, defaultValue int, help string) {
	*param = defaultValue
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: 1,
		DefaultValue:     defaultValue,
		Help:             help,
	})
	// 预留 -c=10 -c10 两种写法的解析
	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				if i+1 < len(args) {
					// 判断下一个参数是否是选项,如果是选项则不解析
					if !(strings.HasPrefix(args[i+1], "-") || strings.HasPrefix(args[i+1], "--")) {
						_, err := fmt.Sscanf(args[i+1], "%d", param)
						if err != nil {
							fmt.Printf(strings.Join([]string{"选项", o, "参数错误", "预期类型:", reflect.TypeOf(param).String()}, " ") + "\n")
							os.Exit(1)
						}
					}
				}
			}
		}
	}
}

func (p *ParamList) Float(options []string, param *float64, defaultValue float64, help string) {
	*param = defaultValue
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: 1,
		DefaultValue:     defaultValue,
		Help:             help,
	})

	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				if i+1 < len(args) {
					if !(strings.HasPrefix(args[i+1], "-") || strings.HasPrefix(args[i+1], "--")) {
						_, err := fmt.Sscanf(args[i+1], "%.2f", param)
						if err != nil {
							fmt.Printf(strings.Join([]string{"选项", o, "参数错误", "预期类型:", reflect.TypeOf(param).String()}, " ") + "\n")
							os.Exit(1)
						}
					}
				}
			}
		}
	}
}

func (p *ParamList) String(options []string, param *string, defaultValue string, help string) {
	*param = defaultValue
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: 1,
		DefaultValue:     defaultValue,
		Help:             help,
	})

	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				if i+1 < len(args) {
					if !(strings.HasPrefix(args[i+1], "-") || strings.HasPrefix(args[i+1], "--")) {
						*param = args[i+1]
					}
				}
			}
		}
	}
}

func (p *ParamList) IpAddr(options []string, param *string, defaultValue string, help string) {
	var ip string
	p.String(options, &ip, defaultValue, help)
	*param = ip
	// 检查IP地址是否合法
	if net.ParseIP(ip) == nil {
		fmt.Printf("IP地址不合法: %s\n", ip)
		os.Exit(1)
	}
}

func (p *ParamList) Path(options []string, param *string, defaultValue string, mustExist bool, help string) {
	args := strings.Join(os.Args, " ")
	if strings.Contains(args, "-h") || strings.Contains(args, "--help") {
		return // 使用帮助时不检查路径
	}
	var path string
	p.String(options, &path, defaultValue, help)
	*param = path
	// 检查路径是否存在
	if !mustExist {
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("路径不存在: %s\n", path)
		os.Exit(1)
	}
}

func (p *ParamList) Ints(options []string, param *[]int, expectParamCount int, defaultValues []int, help string) {
	*param = make([]int, expectParamCount)
	if len(defaultValues) != expectParamCount {
		fmt.Println("默认值的数量必须与期望的参数数量相等")
		os.Exit(1)
	}
	for i := range *param {
		(*param)[i] = defaultValues[i]
	}
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: expectParamCount,
		DefaultValue:     defaultValues,
		Help:             help,
	})

	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				for j := 0; j < expectParamCount; j++ {
					if i+j+1 < len(args) {
						if !(strings.HasPrefix(args[i+j+1], "-") || strings.HasPrefix(args[i+j+1], "--")) {
							_, err := fmt.Sscanf(args[i+j+1], "%d", &(*param)[j])
							if err != nil {
								fmt.Printf(strings.Join([]string{"选项", o, "参数错误", "预期类型:", reflect.TypeOf(param).String()}, " ") + "\n")
								os.Exit(1)
							}
						}
					}
				}
			}
		}
	}
}

func (p *ParamList) Floats(options []string, param *[]float64, expectParamCount int, defaultValues []float64, help string) {
	*param = make([]float64, expectParamCount)
	if len(defaultValues) != expectParamCount {
		fmt.Println("默认值的数量必须与期望的参数数量相等")
		os.Exit(1)
	}
	for i := range *param {
		(*param)[i] = defaultValues[i]
	}
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: expectParamCount,
		DefaultValue:     defaultValues,
		Help:             help,
	})

	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				for j := 0; j < expectParamCount; j++ {
					if i+j+1 < len(args) {
						if !(strings.HasPrefix(args[i+j+1], "-") || strings.HasPrefix(args[i+j+1], "--")) {
							_, err := fmt.Sscanf(args[i+j+1], "%f", &(*param)[j])
							if err != nil {
								fmt.Printf(strings.Join([]string{"选项", o, "参数错误", "预期类型:", reflect.TypeOf(param).String()}, " ") + "\n")
								os.Exit(1)
							}
						}
					}
				}
			}
		}
	}
}

func (p *ParamList) Strings(options []string, param *[]string, expectParamCount int, defaultValues []string, help string) {
	*param = make([]string, expectParamCount)
	if len(defaultValues) != expectParamCount {
		fmt.Println("默认值的数量必须与期望的参数数量相等")
		os.Exit(1)
	}
	for i := range *param {
		(*param)[i] = defaultValues[i]
	}
	p.List = append(p.List, &Args{
		Options:          options,
		ParamType:        *param,
		IsFlag:           false,
		ExpectParamCount: expectParamCount,
		DefaultValue:     defaultValues,
		Help:             help,
	})

	args := os.Args
	for i, v := range args {
		for _, o := range options {
			if v == o {
				for j := 0; j < expectParamCount; j++ {
					if i+j+1 < len(args) {
						if !(strings.HasPrefix(args[i+j+1], "-") || strings.HasPrefix(args[i+j+1], "--")) {
							(*param)[j] = args[i+j+1]
						}
					}
				}
			}
		}
	}
}

func Test() {
	// var count int32
	// count = ParseArgs("-c", int32(0), false, 1, int32(0)).(int32)
	// l.Info(count)
	// fmt.Println("Start")
	// panic("Something went wrong")
	// fmt.Println("End") // 这行代码不会被执行
	// panic(strings.Join([]string{"选项", "-c", "参数错误", "预期类型:", reflect.TypeOf(int(0)).String(), "未指定参数"}, " "))

	// ArgsParse Help
	// Usage: go run . [flag]|[option <args>] ...

	// 	Options     Type            IsFlag          Count   Default         Remark
	// 	-c          int             No              1       10              连接数
	// 	-g          int             No              1       0               分组大小
	// 	-i          int             No              1       0               分组索引
	// 	-h          bool            Yes             0       false           帮助
	// 	-v          bool            Yes             0       false           版本
	// 	-d          string          No              1       GAHDGA          调试模式
	// 	-l          string          No              1       ""              日志级别
	// 	-f          []string        No              5       [a b c]         日志文件
	// 	-h/--help   bool            Yes             0       false           帮助
	parse := ParamList{}
	// args.Help()
	var count int
	var size int
	var index int
	var start bool
	var addr string
	var group []int
	var minmax []float64
	var names []string
	parse.Flag([]string{"-h", "--help"}, &start, false, "帮助")
	parse.Int([]string{"-c", "--count"}, &count, 10, "连接数")
	parse.Int([]string{"-i"}, &index, 0, "分组索引")
	parse.Flag([]string{"-s", "--start"}, &start, false, "开始")
	parse.String([]string{"-addr"}, &addr, "baidu.com", "地址")
	parse.Ints([]string{"-g"}, &group, 2, []int{10, 1}, "分组")
	parse.Floats([]string{"-m", "--min"}, &minmax, 2, []float64{0, 100}, "最小值")
	parse.Strings([]string{"-n", "--name"}, &names, 3, []string{"a", "b", "c"}, "名称")
	parse.CollectArgsInfo("go run .", "0.1.0")

	parse.Help()
	fmt.Println()
	fmt.Println("结果", count, size, index, addr, start, group, minmax, names)
	// go run parserargs.go -c --count 35 -s 43 -i 654 -addr "google.com" -n 我们 他 -m 432543.65 756
}

// go run pargs/pargs.go
// 对比
// go run pargs/pargs.go -c --count 35 -s 43 -i 654 -addr "google.com" -n 我们 他 -m 432543.65 756
// func main() {
// 	Test()
// }
