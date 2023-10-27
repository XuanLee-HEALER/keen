package pvd_test

import (
	"encoding/json"
	"os"
	"testing"

	"gitea.fcdm.top/lixuan/keen/pvd"
	"github.com/cnyjp/fcdmpublic/model"
)

func TestReturnStatus1(t *testing.T) {
	rt := pvd.ReturnStatus{
		Code:    0,
		Message: "",
	}
	rt.Exit()
}

func TestReturnStatus2(t *testing.T) {
	rt := pvd.ReturnStatus{
		Code:    1,
		Message: "common error",
	}
	rt.Exit()
}

func sampleEnv() {
	os.Setenv(model.FCDM_EV_COMMAND, "sample")
	os.Setenv(model.FCDM_EV_APPNAME, "sample_app")
	os.Setenv(model.FCDM_EV_APP_EXTENSION, "sample_extension")
	os.Setenv(model.FCDM_EV_JOBSTEP, "sample_job_step")
	os.Setenv(model.FCDM_EV_JOB_TYPE, "sample_job_type")

	os.Setenv(model.FCDM_EV_AD_PREFIX+"sample_config1", "config1")
	os.Setenv(model.FCDM_EV_AD_PREFIX+"sample_config2", "config2")
	os.Setenv(model.FCDM_EV_IMAGE_AD_PREFIX+"sample_config1", "compat_sample_config1")
	os.Setenv(model.FCDM_EV_IMAGE_AD_PREFIX+"sample_img_config1", "img_config1")
	os.Setenv(model.FCDM_EV_IMAGE_AD_PREFIX+"sample_img_config2", "img_config2")

	os.Setenv(model.FCDM_EV_VOLUME_PREFIX, "sample_volume")
	os.Setenv(model.FCDM_EV_VOLUME_IDENTITY_PREFIX, "sample_volume_identity")
}

func TestNewFCDMArgument(t *testing.T) {
	sampleEnv()
	args := pvd.NewFCDMArgument()
	bs, _ := json.MarshalIndent(args, "", "  ")
	t.Log("\n", string(bs))
}

func TestGetConfig(t *testing.T) {
	sampleEnv()
	args := pvd.NewFCDMArgument()

	v, err := args.GetConfig("sample_config1", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "config1" {
		t.FailNow()
	}
}

func TestGetImgConfig(t *testing.T) {
	sampleEnv()
	args := pvd.NewFCDMArgument()

	v, err := args.GetImgConfig("sample_img_config1", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "img_config1" {
		t.FailNow()
	}
}

func TestGetCompatConfig(t *testing.T) {
	sampleEnv()
	args := pvd.NewFCDMArgument()

	v, err := args.GetCompatConfig("sample_config1", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "compat_sample_config1" {
		t.FailNow()
	}

	v, err = args.GetCompatConfig("sample_config2", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "config2" {
		t.FailNow()
	}

	v, err = args.GetCompatConfig("sample_img_config1", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "img_config1" {
		t.FailNow()
	}

	v, err = args.GetCompatConfig("non_exist", false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if v != "" {
		t.FailNow()
	}

}
