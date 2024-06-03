package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/melbahja/goph"
)

func main() {
	cli, err := goph.NewUnknown("stu095", "pilogin.hpc.sjtu.edu.cn", goph.Password("9YD2LWaz7mzE"))
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	//上传路径下的所有文件
	dirpath := "./func"
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		panic(err)
	}
	slurmName := "/lustre/home/acct-stu/stu095/"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		//如果后缀不是.cu和.slurm,则跳过
		if entry.Name()[len(entry.Name())-3:] != ".cu" && entry.Name()[len(entry.Name())-6:] != ".slurm" {
			continue
		}
		err := cli.Upload(dirpath+"/"+entry.Name(), "/lustre/home/acct-stu/stu095/"+entry.Name())
		if err != nil {
			panic(err)
		}
		//找到后缀为.slurm的文件
		if entry.Name()[len(entry.Name())-6:] == ".slurm" {
			slurmName += entry.Name()
		}

	}
	println(slurmName)
	res, err := cli.Run("sbatch " + slurmName)
	if err != nil {
		panic(err)
	}
	jobname := os.Getenv("JOB_NAME")
	fmt.Printf("Job name: %s\n", jobname)
	jobnamespace := os.Getenv("JOB_NAMESPACE")
	fmt.Printf("Job namespace: %s\n", jobnamespace)
	var jobID string
	fmt.Sscanf(string(res), "Submitted batch job %s", &jobID)
	fmt.Printf("Job ID: %s\n", jobID)
	workdir := "/lustre/home/acct-stu/stu095/"
	ticker := time.NewTicker(5 * time.Second)
	// defer ticker.Stop()
	// 开启一个goroutine执行轮询操作
loop:
	for {
		select {
		case <-ticker.C:
			cmd := "sacct -j " + jobID + " | sed -n '3p' | awk '{print $6}'"
			fmt.Println(cmd)
			res, err = cli.Run(cmd)
			if err != nil {
				fmt.Println(err)
				break loop
			}
			status := string(res)
			//删除status末尾的换行符
			status = strings.TrimSuffix(status, "\n")
			fmt.Printf("Job status: %s\n", status)
			if status == "COMPLETED" {
				fmt.Println("Job completed")
				outputfile := os.Getenv("OUTPUT_FILE")
				// file, err := os.ReadFile(workdir + outputfile)
				err := cli.Download(workdir+outputfile, "/app/"+outputfile)
				if err != nil {
					// panic(err)
					fmt.Println(err)
					break loop
				}
				file, err := os.ReadFile("/app/" + outputfile)
				if err != nil {
					// panic(err)
					fmt.Println(err)
					break loop
				}
				apiserverIp := os.Getenv("API_SERVER_IP")
				apiserverPort := os.Getenv("API_SERVER_PORT")
				fmt.Println(string(file))
				// jsonData := fmt.Sprintf(`{
				// 	"job_id": "%s",
				// 	"output": "%s",
				// 	"jobname": "%s",
				// 	"jobnamespace": "%s"
				// }`, jobID, string(file), jobname, jobnamespace)
				body := map[string]interface{}{
					"job_id":       jobID,
					"output":       string(file),
					"jobname":      jobname,
					"jobnamespace": jobnamespace,
				}
				requestBody, err := json.Marshal(body)
				if err != nil {
					fmt.Println(err)
					break loop
				}
				url := fmt.Sprintf("http://%s:%s/uploadJobOutputResult", apiserverIp, apiserverPort)
				req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
				if err != nil {
					fmt.Println(err)
					break loop
				}
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, er := client.Do(req)
				if er != nil {
					fmt.Println(er)
					break loop
				}
				defer resp.Body.Close()
				break loop
			}
			if status == "FAILED" {
				fmt.Println("Job failed")
				errorfile := os.Getenv("ERROR_FILE")
				// file, err := os.ReadFile(workdir + errorfile)
				err := cli.Download(workdir+errorfile, "/app/"+errorfile)
				if err != nil {
					fmt.Println(err)
					break loop
				}
				file, err := os.ReadFile("/app/" + errorfile)
				if err != nil {
					fmt.Println(err)
					break loop
				}
				apiserverIp := os.Getenv("API_SERVER_IP")
				apiserverPort := os.Getenv("API_SERVER_PORT")
				// jsonData := fmt.Sprintf(`{
				// 	"job_id": "%s",
				// 	"error": "%s",
				// 	"jobname": "%s",
				// 	"jobnamespace": "%s"
				// }`, jobID, string(file), jobname, jobnamespace)
				body := map[string]interface{}{
					"job_id":       jobID,
					"error":        string(file),
					"jobname":      jobname,
					"jobnamespace": jobnamespace,
				}
				requestBody, err := json.Marshal(body)
				if err != nil {
					fmt.Println(err)
					break loop
				}

				url := fmt.Sprintf("http://%s:%s/uploadJobErrorResult", apiserverIp, apiserverPort)
				req, er := http.NewRequest("POST", url, bytes.NewReader(requestBody))
				if er != nil {
					fmt.Println(er)
					break loop
				}
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, er := client.Do(req)
				if er != nil {
					fmt.Println(er)
					break loop
				}
				defer resp.Body.Close()
				break loop
			}
			if status == "CANCELLED" {
				fmt.Println("Job cancelled")
				break
			}
			if status == "PENDING" {
				fmt.Println("Job pending")
			}
		}
	}

	select {}

}
