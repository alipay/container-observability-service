import React, { useState } from 'react';
import { Button, Input, Card, IconButton, Field, Collapse } from '@grafana/ui';
import { StandardEditorProps } from '@grafana/data';
import { FilterConfig } from 'types';

// 用户对数据源的配置面板，可以控制数据是否在圆心展示，是否作为外边框参数，数值类型和单位
export const FilterConfigEditort = ({ item, value, onChange, context }: StandardEditorProps<FilterConfig[]>) => {
    const [isOpen, setIsOpen] = useState(false)

    // 点击Add按钮时添加一行
    const addLine = () => {
        onChange([...value, { filterKey: '', optionConnectMark: '', valueConnectMark: '', keyPrefix: '', keySuffix: '', valuePrefix: '', valueSuffix: '' }])
    }

    // 点击删除图标时移除对应行
    const removeLine = (index: number) => {
        value.splice(index, 1);
        onChange(value)
    }

    const changeInputValue = (inputValue: string, index: number, key: string) => {
        //@ts-ignore
        value[index][key] = inputValue
        onChange(value)
    }

    return <>
        <Collapse label="Filter Config" isOpen={isOpen} onToggle={() => setIsOpen(!isOpen)}>
            {
                Array.isArray(value) && value.map((line, index) => {
                    return <Card key={index}>
                        <Card.Description>
                            <Field label="filter关键字" description='filter的名称, 与constant中的配置一致'>
                                <Input
                                    value={line.filterKey}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'filterKey') }}></Input>
                            </Field>
                            <Field label="键值连接符" description='键值间的连接符, 如配置为=, filter将显示为: key=value' >
                                <Input
                                    value={line.valueConnectMark}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'valueConnectMark') }}></Input>
                            </Field>
                            <Field label="filter连接符" description='每一项之间的连接符, 用来拼接整个filter, 如配置为&, filter将显示为: optionA&optionB'>
                                <Input
                                    value={line.optionConnectMark}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'optionConnectMark') }}></Input>
                            </Field>
                            <Field label="键前缀" description='选项的前缀, 适用于一些特殊格式的配置'>
                                <Input
                                    value={line.keyPrefix}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'keyPrefix') }}></Input>
                            </Field>
                            <Field label="键后缀" description='选项的后缀, 适用于一些特殊格式的配置'>
                                <Input
                                    value={line.keySuffix}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'keySuffix') }}></Input>
                            </Field>
                            <Field label="值前缀" description='选项的前缀, 适用于一些特殊格式的配置'>
                                <Input
                                    value={line.valuePrefix}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'valuePrefix') }}></Input>
                            </Field>
                            <Field label="值后缀" description='选项的后缀, 适用于一些特殊格式的配置'>
                                <Input
                                    value={line.valueSuffix}
                                    //@ts-ignore
                                    onChange={(e) => { changeInputValue(e.target.value, index, 'valueSuffix') }}></Input>
                            </Field>
                        </Card.Description>
                        <Card.SecondaryActions>
                            <IconButton
                                key="delete"
                                name="trash-alt"
                                onClick={() => { removeLine(index) }}
                                tooltip="Delete this config" />
                        </Card.SecondaryActions>
                    </Card>
                })
            }
            <Button icon="plus" onClick={() => addLine()}>Add Filter</Button>
        </Collapse>
    </>
};

