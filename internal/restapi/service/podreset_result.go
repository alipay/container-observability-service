package service

import (
    "github.com/alipay/container-observability-service/internal/restapi/model"
    storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertResetResult2Frame(responseResult storagemodel.Response) []model.ResetResult {
    bit := model.ResetResult{
        Code: responseResult.Code,
        Message:   responseResult.Message,
        Status:   responseResult.Status,
        Data:   responseResult.Data,
    }

    return []model.ResetResult{bit}
}