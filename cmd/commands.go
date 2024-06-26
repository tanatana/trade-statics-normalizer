package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// RootCmd defines root command
	rootCmd = &cobra.Command{
		Use:   "trade-statics-normalizer",
		Short: "A simple normalizer for Japan custom trade statics",
		Long: `trade-statics-normalizer is a CLI tool for normalizing Japan custom trade statics data.
The trade statics data published from Japan custom is so much simple CSV file.
So if you want to use them, almost of all usecases,
you need normalize these data. Then this tool helps you.`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Usage()
				Exit(nil, 1)
			}

			filters := strings.Split(filter, ",")

			newRows := [][]string{}
			newRows = append(newRows, []string{"expOrImp", "year", "month", "HS", "data category", "unit", "value"})
			for _, arg := range args {
				file, err := os.Open(arg)
				if err != nil {
					Exit(err, 1)
				}
				defer file.Close()

				r := csv.NewReader(file)
				rows, err := r.ReadAll()
				if err != nil {
					Exit(err, 1)
				}
				for rowIdx, row := range rows {
					if rowIdx == 0 {
						continue
					}
					expOrImp := row[0]
					year := row[1]
					HS := row[2]
					// HS コードのクォートを取り除く
					HS = HS[1 : len(HS)-1]
					if filter != "" && len(filters) > 0 && !slices.Contains(filters, HS) {
						continue
					}

					for _, col := range []int{3, 4, 5} {
						unit := row[col]
						cat := ""
						switch col {
						case 3:
							cat = "Quantity1"
						case 4:
							cat = "Quantity2"
						case 5:
							cat = "Volume"
							unit = ""
						}

						// ignore empty unit
						if unit == "  " {
							continue
						}
						for _, month := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} {
							monthStr := fmt.Sprintf("%d", month)
							// Unit1 は 3列目、 Unit2 は4列目 なので
							// col から 3 を引いてオフセットを求める
							offset := col - 3
							// データは 5列目から始まる。
							// 最初に年のデータが3列あるので、1月のデータは 1*3 で
							// アクセス可能
							value := row[5+offset+month*3]

							newRow := []string{expOrImp, year, monthStr, HS, cat, unit, value}
							newRows = append(newRows, newRow)
						}
					}
				}
			}

			var w *csv.Writer
			if out == "" {
				w = csv.NewWriter(os.Stdout)
			} else {
				file, err := os.Create(out)
				if err != nil {
					Exit(err, 1)
				}
				defer file.Close()
				w = csv.NewWriter(file)
			}
			defer w.Flush()

			err := w.WriteAll(newRows)
			if err != nil {
				Exit(err, 1)
			}
		},
	}
	filter string
	out    string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&out, "out", "o", "", "Output file name. Default is STDOUT.")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "HS code filter. Default is all.")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// Exit finishes a runnning action.
func Exit(err error, codes ...int) {
	var code int
	if len(codes) > 0 {
		code = codes[0]
	} else {
		code = 2
	}
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
