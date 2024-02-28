/**
 * @description 使用时必须于请求头加上responseType: 'blob'，慎用，慎用！
 * @stream  下载流（格式）
 * @filename  下载文件名称
 * @suffix  下载文件后缀，例如".xlsx"
 * @stream: any  等参数格式是对参数类型做一下限制
 */
 
function downloadFile(
  stream: any,                         
  filename: string,  //不传值默认以当前时间为文件名
  suffix: string,
) {
  //通过new Blob和文件格式生成blob对象
  const blob = new Blob([stream]);
  const objectURL = URL.createObjectURL(blob);
  let link = document.createElement('a');
  //下载的文件名
  link.download = `${filename}${suffix}`; 
  link.href = objectURL;
  // @ts-ignore 判断浏览器的方法
  if (!!window.ActiveXObject || 'ActiveXObject' in window) {
    // @ts-ignore 判断浏览器的方法
    window.navigator.msSaveOrOpenBlob(blob, filename);
  } else {
    link.click();
  }
  URL.revokeObjectURL(objectURL);
  // @ts-ignore
  link = null;
}
 
export default downloadFile;

