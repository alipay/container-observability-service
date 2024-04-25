import React, { useMemo } from 'react';
import { SceneFlexLayout, EmbeddedScene, SceneFlexItem, PanelBuilders, SceneQueryRunner } from '@grafana/scenes';
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';
import './SimplePanel.css'

interface Props extends PanelProps<SimpleOptions> { }

// create log panel with query and options
function getSceneItem(query: SceneQueryRunner, options: SimpleOptions): SceneFlexItem {
  return new SceneFlexItem({
    width: '100%',
    height: '100%',
    //@ts-ignore
    body: PanelBuilders.logs()
      .setData(query)
      .setOption("dedupStrategy", options.dedupStrategy)
      .setOption("showTime", options.showTime)
      .setOption("enableLogDetails", options.enableLogDetails)
      .setOption("prettifyLogMessage", options.prettifyLogMessage)
      .setOption("showCommonLabels", options.showCommonLabels)
      .setOption("showLabels", options.showLabels)
      .setOption("sortOrder", options.sortOrder)
      .setOption("wrapLogMessage", options.wrapLogMessage)
      .build(),
  })
}

function formatUrl(urlStr: string, paramKey: string, paramValue: string): string {
  paramKey = paramKey.replace('var-', '');
  return urlStr.replace(`$${paramKey}`, paramValue)
}

function getUrl(urlUser: string, urlPanel: string): string {
  const param = urlPanel.split('?')[1]
  if (urlUser === '') {
    return urlPanel;
  } else if (urlUser.indexOf('$param') > -1) {
    return urlUser.replace(/\$param/g, param)
  } else {
    return urlUser
  }
}

//Function stabilization
function debounce(fn: any, wait: number) {
  let timer: any = null;
  return function () {
    if (timer !== null) {
      clearTimeout(timer);
    }
    timer = setTimeout(fn, wait);
  }
}

export const SimplePanel: React.FC<Props> = ({ options, data, width, height, replaceVariables }) => {
  const params: any = {}
  options.params.map(param =>{
    if (document.getElementById(`var-${param}`)?.innerText) {
      params[param] = document.getElementById(`var-${param}`)?.innerText
    } else {
      //@ts-ignore
      params[param] = document.getElementById(`var-${param}`)?.value
    }
  })

  data.request?.targets.map(target => {
    if (target.datasource?.type === 'yesoreyeram-infinity-datasource') {
      //@ts-ignore
      let url: string = target.url
      Object.keys(params).map(key => {
        //@ts-ignore
        url = formatUrl(url, key, params[key])
      })
      //@ts-ignore
      target.url = url
      return target
    } else {
      return target
    }
  })

  const query = new SceneQueryRunner({
    //@ts-ignore
    queries: data.request?.targets,
  })

  const sceneItem = getSceneItem(query, options)
  const scene = useMemo(() => {
    return new EmbeddedScene({
      body: new SceneFlexLayout({
        children: [
          sceneItem
        ],
      }),
    })
  }, [sceneItem])

  // observer DOM which han been changed.
  function callback(mutationsList: any, observer: any) {
    const tdElement = document.querySelectorAll('td')
    const fieldConfig = options.label
    for (let mutation of mutationsList) {
      if (mutation.type === 'childList') {
        fieldConfig.map(config => {
          if (config.name === '') {
            return
          }
          tdElement.forEach(element => {
            if (element.innerText && (element.innerText.indexOf(config.name) === 0)) {
              const target = element.nextSibling?.firstChild as HTMLDivElement
              if (!target.id) {
                switch (config.type) {
                  case 'dataLink':
                    target.id = getUrl(config.url, target.innerText)
                    target.innerHTML = config.overrideName === '' ? target.innerText : config.overrideName
                    target.setAttribute('class', 'link')
                    target.addEventListener('click', (e) => {
                      window.open(target.id, '_blank');
                    })
                    break;
                  case 'text':
                    target.innerHTML = config.overrideName;
                    break;
                  default:
                    break;
                }
              }
            }
          })
        })
      }
    }
  }

  sceneItem.state.body?.addActivationHandler(() => {
    setTimeout(() => {
      const tableElement = document.querySelectorAll('tbody');
      tableElement.forEach(element => {
        // 创建一个观察者对象，并传入回调函数
        const observer = new MutationObserver(callback);
        // 观察者的配置（观察目标节点的子节点的变化）
        const config = { childList: true };
        // 开始观察目标节点
        debounce(observer.observe(element, config), 200);
      })
    }, 1000)

  })
  return <scene.Component model={scene} />
};

