package main

import  (
	"fmt"
	flag "github.com/spf13/pflag"
	//"flag"
	"strings"
	"os"
	"os/exec"
	"bufio"
	"io"
)

type sp_args struct{
	start_page int
	end_page int
	in_filename string
	page_len int
	page_type bool
	print_dest string
}

func main()  {
	sa:=new(sp_args)

	//定义标志参数
	flag.IntVarP(&sa.start_page,"start","s",-1,"the start page")
	flag.IntVarP(&sa.end_page,"end","e",-1,"the end page")
	flag.IntVarP(&sa.page_len,"length","l",72,"the page length")
	flag.StringVarP(&sa.print_dest,"dest","d","","the printer")

	//检查 -f是否存在，注意 -f 只支持bool类型
	flag.BoolVarP(&sa.page_type,"type","f",false,"end with EOF")
	//解析命令行参数到定义的flag
	flag.Parse()
	
	// 获取non-flag参数
	if len(flag.Args()) == 1 {
		sa.in_filename = flag.Args()[0]
	}else{
		sa.in_filename=""
	}

	process_args(*sa,flag.NArg())
	process_input(*sa)
}

func usage(){
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg [-s start_page] [-e end_page] [ -l lines_per_page | -f ] [ -d dest ] [ in_filename ]\n")
}

func process_args(sa sp_args, nonFlagNum int){
	s_e_ok := sa.start_page <= sa.end_page && sa.start_page >= 1
    num_ok := nonFlagNum == 1 || nonFlagNum == 0
    l_f_ok := sa.page_type && sa.page_len != 72
    if !s_e_ok || !num_ok || l_f_ok {
        usage()
        os.Exit(1)
    }
}

func process_input(sa sp_args){
	currPage := 1
    currLine := 0

    fin := os.Stdin
    fout := os.Stdout
    var inpipe io.WriteCloser
	var err error
	
	//确定输入源，是否改为从文件读入
	if sa.in_filename!="" {
		fin, err = os.Open(sa.in_filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selpg: could not open input file \"%s\"\n", sa.in_filename)
            fmt.Println(err)
            usage()
            os.Exit(1)
		}
		defer fin.Close()
	}

	/**确定输出源, 是否打印到打印机
     * 由于没有打印机测试，用管道接通 grep 作为测试，结果输出到屏幕
     * selpg内容通过管道输入给 grep, grep从中搜出带有keyword文件的内容
	 */
    if sa.print_dest != "" {
        cmd := exec.Command("grep", "-nf", "keyword")
        inpipe, err = cmd.StdinPipe()
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer inpipe.Close()
        cmd.Stdout = fout
        cmd.Start()
    }

	// 按行分页读取输出
	if !sa.page_type {
		line := bufio.NewScanner(fin)
		for line.Scan() {
			if currPage >= sa.start_page && currPage <= sa.end_page {
				if sa.print_dest != "" {
					inpipe.Write([]byte(line.Text() + "\n"))
				}else {
					fout.Write([]byte(line.Text() + "\n"))
				}
			 }
			 currLine++
			 if currLine%sa.page_len == 0 {
				 currPage++
				 currLine = 0
			 }
		}
	}	else { // 按分隔符分页读取输出
		rd := bufio.NewReader(fin)
		for {
			page, ferr := rd.ReadString('\f')
			if ferr != nil { // 出错
				if ferr == io.EOF {
					if currPage >= sa.start_page && currPage <= sa.end_page {
						fmt.Fprintf(fout, "%s", page)
					}
				}
				break
			}
			page = strings.Replace(page,"\f","",-1)
			if currPage >= sa.start_page && currPage <= sa.end_page {
				fmt.Println(currPage)
				fmt.Fprintf(fout,"%s",page)
				currPage++			
			}else{
				break
			}
			
		}
		if currPage < sa.end_page {
			fmt.Fprintf(os.Stderr, "./selpg: end_page (%d) greater than total pages (%d), less output than expected\n", sa.end_page, currPage)
		}
	}
}