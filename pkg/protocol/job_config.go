package protocol

type Job struct {
	ApiVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   MetadataType `yaml:"metadata" json:"metadata"`
	Spec       JobSpec      `yaml:"spec" json:"spec"`
	Status     JobStatus    `yaml:"status" json:"status"`
}

type JobSpec struct {
	Partition      string   `yaml:"partition" json:"partition"`
	OutputFile     string   `yaml:"outputFile" json:"outputFile"`
	ErrorFile      string   `yaml:"errorFile" json:"errorFile"`
	RunCommands    []string `yaml:"runCommands" json:"runCommands"`
	GPUNums        string   `yaml:"GPUNums" json:"GPUNums"`
	CPUPerTask     string   `yaml:"CPUPerTask" json:"CPUPerTask"`
	UploadPath     string   `yaml:"uploadPath" json:"uploadPath"`
	NTasks         string   `yaml:"nTasks" json:"nTasks"`
	NTasksPerNode  string   `yaml:"nTasksPerNode" json:"nTasksPerNode"`
	UserUploadFile []byte   `yaml:"userUploadFile" json:"userUploadFile"`
}

type JobStatus struct {
	StartTime         string `yaml:"startTime" json:"startTime"`
	JobState          string `yaml:"jobState" json:"jobState"`
	OutputFileContent string `yaml:"outputFileContent" json:"outputFileContent"`
	ErrorFileContent  string `yaml:"errorFileContent" json:"errorFileContent"`
}
