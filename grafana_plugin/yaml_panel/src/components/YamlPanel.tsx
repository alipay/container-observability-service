import React, { useEffect, useRef, useState } from 'react';
import axios from 'axios';
import { DownloadOutlined } from '@ant-design/icons';
import { Button, ConfigProviderProps, Alert } from 'antd';
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';
import { UnControlled as CodeMirror } from 'react-codemirror2';
import 'codemirror/lib/codemirror.css';
import 'codemirror/lib/codemirror.js';
import 'codemirror/mode/yaml/yaml';
import 'codemirror/mode/javascript/javascript';
import 'codemirror/theme/base16-dark.css'
import 'codemirror/theme/idea.css'
//ctrl+空格代码提示补全
import 'codemirror/addon/hint/show-hint.css';
import 'codemirror/addon/hint/show-hint';
import 'codemirror/addon/lint/lint';
import 'codemirror/addon/lint/json-lint';
import 'codemirror/addon/hint/anyword-hint.js';
//代码高亮
import 'codemirror/addon/selection/active-line';
//折叠代码
import 'codemirror/addon/fold/foldgutter.css';
import 'codemirror/addon/fold/foldcode.js';
import 'codemirror/addon/fold/foldgutter.js';
import 'codemirror/addon/fold/brace-fold.js';
import 'codemirror/addon/fold/xml-fold.js';
import 'codemirror/addon/fold/indent-fold.js';
import 'codemirror/addon/fold/markdown-fold.js';
import 'codemirror/addon/fold/comment-fold.js';
import 'codemirror/addon/edit/closebrackets';
import './YamlPanel.css';
import downloadFile from '../util/download';
import {DisplayModel, Theme} from '../types';

const yaml = require('json2yaml')
const size: SizeType = 'middle'
const searchParams = new URLSearchParams(window.location.search)
const JsonModel = {
  name: "javascript",
  json: true
}
const YamlModel = {
  name: 'text/x-yaml'
}
const ParamsKind = ['uid', 'name', 'hostname', 'podip']

type SizeType = ConfigProviderProps['componentSize'];
interface AlertState {
  visible: boolean;
  type: 'warning' | 'error';
  message: 'Warning' | 'Error';
  description: string;
}

interface Props extends PanelProps<SimpleOptions> { }

export const SimplePanel: React.FC<Props> = ({options, data}) => {
  const cmRef = useRef(null);
  const [yamlString, setYamlString] = useState('')
  const [theme, setTheme] = useState(options.theme)
  const [alertState, setAlertState] = useState<AlertState>({ visible: false, type: 'warning', message: 'Warning', description: '' })
  const [model, setModel] = useState(options.displayModel)
  const [params, setParams] = useState({resourece: '', type: '', value: ''})

  const changeTheme = (theme: Theme) => {
    setTheme(theme)
  }
  

  const changeModel = (newModel: DisplayModel) => {
      //@ts-ignore
      const cm = cmRef.current.editor
      if (newModel === model) {
        return
      }
      if(newModel === 'yaml') {
        cm.setOption("mode", YamlModel)
      } else {
        cm.setOption("mode", JsonModel)
      }
      setModel(newModel)
  }

  useEffect(() => {
    const resource = searchParams.get('resource');
    let [paramType, paramValue] = ['', '']
    
    const setValue = (result: string) => {
      if (!result) {
        setYamlString('null');
        return  
      }
      if (model === 'yaml') {
        setYamlString(yaml.stringify(result).replace(/\\"/g, '"'))
      } else {
        setYamlString(JSON.stringify(result, null, 4).replace(/\\"/g, '"'))
      }
    }

    for(let param of ParamsKind) {
      if(searchParams.get(param)){
        paramType = param
        paramValue = searchParams.get(param) as string
        setParams({resourece: resource as string, type: paramType, value: paramValue})
      }
    }
    
    if (resource && paramValue) {
      axios.get(`http://lunettesdi.hcs.svc.alipay.net:18883/queryyamls?resource=${resource}&${paramType}=${paramValue}`)
        .then((response) => {
          let result = ''
          if (response && response.data) {
            result = JSON.parse(response.data.split("var data = ")[1]?.split("var tree = ")[0])
          } else {
            setAlertState({
              visible: true,
              type: 'warning',
              message: 'Warning',
              description: `No yaml responsed.`
            })
          }
          setValue(result)
        })
        .catch((error) => {
          setAlertState({
            visible: true,
            type: 'error',
            message: 'Error',
            description: error.toString()
          })
        })
    } else {
      if (data.state === "Done") {
        const result = data.series[0].meta?.custom?.data
        setValue(result)
      }
    }
  },[options, data, model])

  return (
    <div className='main'>
      {alertState.visible && (
        <Alert message={alertState.message} type={alertState.type} description={alertState.description} closable />
      )}
      <Button type={theme === "idea" ? "primary" : "default"} size={size} onClick={() => changeTheme("idea")}>Light</Button>
      <Button type={theme === "base16-dark" ? "primary" : "default"} size={size} onClick={() => changeTheme("base16-dark")}>Dark</Button>
      <Button type={model === "json" ? "primary" : "default"} size={size} onClick={() => changeModel('json')}>Json</Button>
      <Button type={model === "yaml" ? "primary" : "default"} size={size} onClick={() => changeModel('yaml')}>Yaml</Button>
      <Button type="default" icon={<DownloadOutlined />} size={size} onClick={() => { downloadFile(yamlString, `${params.resourece}_${params.type}_${params.value}`, `.${model}`) }}>
        Download
      </Button>

      <CodeMirror
        ref={cmRef}
        options={{
          styleActiveLine: true,//光标代码高亮
          readOnly: true, // 只读
          lineNumbers: true, // 显示行号
          theme: theme, // 设置主题
          mode: {
            name: 'text/x-yaml', // "text/css" ...
          },
          extraKeys: { "Ctrl-Q": function (cm: any) { cm.foldCode(cm.getCursor()); } },

          // (以下三行)设置支持代码折叠
          lineWrapping: true,
          foldGutter: true,
          gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter'],
        }}
        value={yamlString}

      />
    </div>
  );
};
