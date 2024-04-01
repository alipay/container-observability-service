import React, { useEffect, useRef, useState } from 'react';
import { DownloadOutlined } from '@ant-design/icons';
import { Button, ConfigProviderProps } from 'antd';
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
import { DisplayModel, Theme } from '../types';

const yaml = require('json2yaml')
const size: SizeType = 'middle'
const JsonModel = {
  name: "javascript",
  json: true
}
const YamlModel = {
  name: 'text/x-yaml'
}

type SizeType = ConfigProviderProps['componentSize'];
interface Props extends PanelProps<SimpleOptions> { }
export const SimplePanel: React.FC<Props> = ({ options, data, width, height }) => {
  const cmRef = useRef(null);
  const [yamlString, setYamlString] = useState('')
  const [theme, setTheme] = useState(options.theme)
  const [model, setModel] = useState(options.displayModel)

  const changeTheme = (theme: Theme) => {
    setTheme(theme)
  }

  const changeModel = (newModel: DisplayModel) => {
    //@ts-ignore
    const cm = cmRef.current.editor
    if (newModel === model) {
      return
    }
    if (newModel === 'yaml') {
      cm.setOption("mode", YamlModel)
    } else {
      cm.setOption("mode", JsonModel)
    }
    setModel(newModel)
  }

  useEffect(() => {
    const setValue = (result: string) => {
      if (model === 'yaml') {
        setYamlString(yaml.stringify(result))
      } else {
        setYamlString(JSON.stringify(result, null, 4))
      }
    }
    //@ts-ignore
    const cm = cmRef.current.editor
    cm.setSize(null, height - 30);
    if (data.state === "Done") {
      const result = data.series[0].meta?.custom?.data
      setValue(result)
    }
  }, [options, data, model, height])

  return (
    <div className='main'>
      <Button type={theme === "idea" ? "primary" : "default"} size={size} onClick={() => changeTheme("idea")}>Light</Button>
      <Button type={theme === "base16-dark" ? "primary" : "default"} size={size} onClick={() => changeTheme("base16-dark")}>Dark</Button>
      <Button type={model === "json" ? "primary" : "default"} size={size} onClick={() => changeModel('json')}>Json</Button>
      <Button type={model === "yaml" ? "primary" : "default"} size={size} onClick={() => changeModel('yaml')}>Yaml</Button>
      <Button type="default" icon={<DownloadOutlined />} size={size} onClick={() => { downloadFile(yamlString, `${new Date().toLocaleDateString()}`, `.${model}`) }}>
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
          // (以下三行)设置支持代码折叠
          lineWrapping: true,
          viewportMargin: 5000,
          foldGutter: true,
          gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter'],
        }}
        value={yamlString}

      />
    </div>
  );
};
