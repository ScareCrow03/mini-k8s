package handler

import (
	"encoding/json"
	"fmt"
	"mini-k8s/pkg/constant"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/protocol"
	"mini-k8s/pkg/utils/uid"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateJob(c *gin.Context) {
	var job protocol.Job
	c.BindJSON(&job)
	job.Metadata.UID = "mini-k8s-job-" + uid.NewUid()
	job.Status.JobState = "Pending"
	job.Status.StartTime = time.Now().Format("2006-01-02 15:04:05")
	if job.Metadata.Namespace == "" {
		job.Metadata.Namespace = "default"
	}
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	defer st.Close()
	jsonstr, err := json.Marshal(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	st.Put(constant.EtcdJobPrefix+job.Metadata.Namespace+"/"+job.Metadata.Name, jsonstr)
	c.JSON(http.StatusOK, gin.H{
		"message": "create job: " + job.Metadata.Namespace + "/" + job.Metadata.Name,
	})
}

func UploadJobOutputResult(c *gin.Context) {
	var res map[string]interface{}
	c.BindJSON(&res)
	fmt.Println(res)
	name := res["jobname"].(string)
	namespace := res["jobnamespace"].(string)
	outputcontent := res["output"].(string)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	defer st.Close()
	job, err := st.Get(constant.EtcdJobPrefix + namespace + "/" + name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if len(job.Value) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "job not found",
		})
		return
	}

	var jobobj protocol.Job
	err = json.Unmarshal(job.Value, &jobobj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	jobobj.Status.OutputFileContent = outputcontent
	jobobj.Status.JobState = "Finished"
	jsonstr, err := json.Marshal(jobobj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	st.Put(constant.EtcdJobPrefix+namespace+"/"+name, jsonstr)
}

func UploadJobErrorResult(c *gin.Context) {
	var res map[string]interface{}
	c.BindJSON(&res)
	name := res["jobname"].(string)
	namespace := res["namespace"].(string)
	errorcontent := res["error"].(string)
	st, err := etcd.NewEtcdStore(constant.EtcdIpPortInTestEnvDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	defer st.Close()
	job, err := st.Get(constant.EtcdJobPrefix + namespace + "/" + name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(job.Value) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "job not found",
		})
		return
	}

	var jobobj protocol.Job
	err = json.Unmarshal(job.Value, &jobobj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	jobobj.Status.ErrorFileContent = errorcontent
	jobobj.Status.JobState = "Error"
	jsonstr, err := json.Marshal(jobobj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	st.Put(constant.EtcdJobPrefix+namespace+"/"+name, jsonstr)
}
