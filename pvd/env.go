package pvd

import (
	"encoding/base64"
	"os"
	"strconv"
	"strings"

	"gitea.fcdm.top/lixuan/keen"
	"github.com/cnyjp/fcdmpublic/model"
)

type ConfigValidator func(FCDMArgument) (string, bool)

type Decoder func(string) (string, error)

var (
	Base64Decoder Decoder = func(s string) (string, error) {
		ds, err := base64.StdEncoding.DecodeString(s)
		return string(ds), err
	}
)

type FCDMArgument struct {
	Command                 string            `json:"command"`
	ApplicationName         string            `json:"application_name"`
	ApplicationExtension    string            `json:"application_extension"`
	Configs                 map[string]string `json:"configs"`
	ImageConfigs            map[string]string `json:"image_configs"`
	VolumeInformation       map[string]string `json:"volumes_information"`
	VolsIdentityInformation map[string]string `json:"volumes_identity_information"`
	BackupType              string            `json:"backup_type"`
	BackupClusterMessage    string            `json:"backup_cluster_message"`
	JobStep                 string            `json:"job_step"`
	JobType                 string            `json:"job_type"`
	JobID                   string            `json:"job_id"` // task id
}

func NewFCDMArgument() FCDMArgument {
	env := make(map[string]string)
	for _, line := range os.Environ() {
		segs := strings.Split(line, "=")
		env[segs[0]] = segs[1]
	}

	cmd := env[model.FCDM_EV_COMMAND]
	appName := env[model.FCDM_EV_APPNAME]
	appExt := env[model.FCDM_EV_APP_EXTENSION]
	jobStep := env[model.FCDM_EV_JOBSTEP]
	jobType := env[model.FCDM_EV_JOB_TYPE]
	jobID := env[model.FCDM_EV_MAINJOB_ID]

	cfg := make(map[string]string)
	icfg := make(map[string]string)
	vi := make(map[string]string)
	vii := make(map[string]string)
	for k, v := range env {
		if strings.HasPrefix(k, model.FCDM_EV_AD_PREFIX) {
			cfg[k] = v
		} else if strings.HasPrefix(k, model.FCDM_EV_IMAGE_AD_PREFIX) {
			icfg[k] = v
		} else if strings.HasPrefix(k, model.FCDM_EV_VOLUME_PREFIX) {
			if strings.HasPrefix(k, model.FCDM_EV_VOLUME_IDENTITY_PREFIX) {
				vii[k] = v
			} else {
				vi[k] = v
			}
		}
	}

	bkt := env[model.FCDM_EV_JOB_BACKUP_TYPE]
	bkc := env[model.FCDM_EV_JOB_INIT_MESSAGE]

	return FCDMArgument{
		cmd, appName, appExt, cfg, icfg, vi, vii, bkt, bkc, jobStep, jobType, jobID,
	}
}

// MapCommand 文件日志名称中命令的对应名称
func (arg FCDMArgument) MapCommand() string {
	switch arg.Command {
	case model.CMD_DISCOVER:
		return "discover"
	case model.CMD_APPLICATION_INFO:
		return "refresh"
	case model.CMD_BACKUP:
		return "backup"
	case model.CMD_MOUNT:
		return "mount"
	case model.CMD_UMOUNT:
		return "umount"
	case model.CMD_RESTORE:
		return "restore"
	case model.CMD_PLUGIN_INFO:
		return "pluginfo"
	default:
		return "illegal"
	}
}

// IsLegal 判断命令是否有效
func (arg FCDMArgument) IsLegal() bool {
	return arg.Command == model.CMD_DISCOVER ||
		arg.Command == model.CMD_APPLICATION_INFO ||
		arg.Command == model.CMD_BACKUP ||
		arg.Command == model.CMD_MOUNT ||
		arg.Command == model.CMD_UMOUNT ||
		arg.Command == model.CMD_RESTORE ||
		arg.Command == model.CMD_PLUGIN_INFO
}

func (arg FCDMArgument) IsSyncDistributeInstance() bool {
	return arg.JobType == model.JOB_TYPE_BACKUP && arg.JobStep == model.JOB_STEP_INIT
}

// GetConfig 获取配置项的值，如果是非编码值可以忽略错误
func (arg FCDMArgument) GetConfig(configName string, isEncode bool, decode Decoder) (string, error) {
	v := arg.Configs[model.FCDM_EV_AD_PREFIX+configName]
	if isEncode {
		return decode(v)
	}
	return v, nil
}

// GetImgConfig 获取镜像配置项的值，如果是非编码值可以忽略错误
func (arg FCDMArgument) GetImgConfig(imageConfigName string, isEncode bool, decode Decoder) (string, error) {
	v := arg.ImageConfigs[model.FCDM_EV_IMAGE_AD_PREFIX+imageConfigName]
	if isEncode {
		return decode(v)
	}
	return v, nil
}

// GetCompatConfig 获取镜像配置覆盖普通配置之后的配置项的值，只覆盖有效值值，即镜像配置不为空字符串的值，如果是非编码值可以忽略错误
func (arg FCDMArgument) GetCompatConfig(configName string, isEncode bool, decode Decoder) (string, error) {
	v, err := arg.GetConfig(configName, isEncode, decode)
	if err != nil {
		return v, err
	}
	sv, err := arg.GetImgConfig(configName, isEncode, decode)
	if sv != "" && err == nil {
		return sv, nil
	}
	return v, err
}

// GetVolume 获取备份设备的目录
func (arg FCDMArgument) GetVolume(volumeIdentity string) string {
	return arg.VolumeInformation[model.FCDM_EV_VOLUME_PREFIX+volumeIdentity]
}

// Validate 对参数的有效值进行基本校验
func (arg FCDMArgument) Validate() bool {
	var r bool
	// 命令必须有效
	r = arg.IsLegal()
	if !r {
		keen.Log.Warn("command is illegal")
		return r
	}

	// job ID 不能为空
	r = arg.JobID != ""
	if !r {
		keen.Log.Warn("the job ID is empty")
		return r
	}

	validVols := func() bool {
		return len(arg.VolumeInformation) > 0
	}

	// 对于备份、恢复、挂载、卸载，volume路径必须存在
	if arg.Command == model.CMD_BACKUP {
		r = validVols()
		if !r {
			keen.Log.Warn("backup: volume information is empty")
			return r
		}

		bt, err := strconv.Atoi(arg.BackupType)
		if err != nil {
			r = false
			keen.Log.Warn("backup: backup type [%s] is illegal: %v", arg.BackupType, err)
			return r
		}
		if bt != model.BACKUP_TYPE_ALL && bt != model.BACKUP_TYPE_DB && bt != model.BACKUP_TYPE_LOG {
			r = false
			keen.Log.Error("backup: backup type [%d] is out of scope", bt)
			return r
		}
	} else if arg.Command == model.CMD_RESTORE {
		r = validVols()
		if !r {
			keen.Log.Warn("restore: volume information is empty")
			return r
		}
	} else if arg.Command == model.CMD_MOUNT {
		r = validVols()
		if !r {
			keen.Log.Warn("mount: volume information is empty")
			return r
		}
	} else if arg.Command == model.CMD_UMOUNT {
		r = validVols()
		if !r {
			keen.Log.Warn("unmount: volume information is empty")
			return r
		}
	}

	return r
}
