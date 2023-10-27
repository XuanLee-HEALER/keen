package pvd

import (
	"os"
	"strings"

	"github.com/cnyjp/fcdmpublic/model"
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
		cmd, appName, appExt, cfg, icfg, vi, vii, bkt, bkc, jobStep, jobType,
	}
}

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
func (arg FCDMArgument) GetConfig(configName string, isEncode bool, decode func(str string) (string, error)) (string, error) {
	v := arg.Configs[model.FCDM_EV_AD_PREFIX+configName]
	if isEncode {
		return decode(v)
	}
	return v, nil
}

// GetImgConfig 获取镜像配置项的值，如果是非编码值可以忽略错误
func (arg FCDMArgument) GetImgConfig(imageConfigName string, isEncode bool, decode func(str string) (string, error)) (string, error) {
	v := arg.ImageConfigs[model.FCDM_EV_IMAGE_AD_PREFIX+imageConfigName]
	if isEncode {
		return decode(v)
	}
	return v, nil
}

// GetCompatConfig 获取镜像配置覆盖普通配置之后的配置项的值，只覆盖有效值值，即镜像配置不为空字符串的值，如果是非编码值可以忽略错误
func (arg FCDMArgument) GetCompatConfig(configName string, isEncode bool, decode func(str string) (string, error)) (string, error) {
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
