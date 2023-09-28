/*
Package podphase 记录一个Pod上所有的操作

分析审计日志，还原成针对Pod的操作
数据写入elasticsearch，索引pod_life_phase
*/
package reason

const NewReasonFeature = "NewReasonFeature"
