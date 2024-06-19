import React, { useState } from 'react';
import { Button, Input, Card, IconButton, Field, Collapse, Select } from '@grafana/ui';
import { SelectableValue, StandardEditorProps } from '@grafana/data';
import { FilterOption } from 'types';

// 用户对数据源的配置面板，可以控制数据是否在圆心展示，是否作为外边框参数，数值类型和单位
export const FilterOptionEditort = ({ item, value, onChange, context }: StandardEditorProps<FilterOption[]>) => {
    const [isOpen, setIsOpen] = useState(false)
    const selectOptions: SelectableValue[] = []

    // 点击Add按钮时添加一行
    const addLine = () => {
        onChange([...value, { label: '', value: '', belongTo: [], isOpen: true }])
    }

    const removeOptions = (index: number) => {
        value.splice(index, 1)
        onChange(value)
    }

    const changeOptionsOpenState = (index: number, state: boolean) => {
        value[index].isOpen = state
        onChange(value)
    }

    const changeInputValue = (inputValue: string, index: number, key: string, optionIndex?: number) => {
        //@ts-ignore
        value[index][key] = inputValue
        onChange(value)
    }

    if (context.options && context.options.filterConfig.length > 0) {
        context.options.filterConfig.map((config: any) => {
            selectOptions.push({label: config.filterKey, value: config.filterKey})
        })
    }

    return <>
        <Collapse label="Filter Config" isOpen={isOpen} onToggle={() => setIsOpen(!isOpen)}>
            {
                Array.isArray(value) && value.map((option, index) => {
                    return <Collapse key={'c' + index} label={option.label} isOpen={option.isOpen} onToggle={() => changeOptionsOpenState(index, !option.isOpen)}>
                        <Card key={index}>
                            <Card.Description>
                                <Field label="Label" description='filter在画面显示的文本'>
                                    <Input
                                        value={option.label}
                                        //@ts-ignore
                                        onChange={(e) => { changeInputValue(e.target.value, index, 'label') }}></Input>
                                </Field>
                                <Field label="Value" description='filter的实际字段' >
                                    <Input
                                        value={option.value}
                                        //@ts-ignore
                                        onChange={(e) => { changeInputValue(e.target.value, index, 'value') }}></Input>
                                </Field>
                                <Field label="BelongTo" description='一个option可以从属于多个filter' >
                                    <Select
                                        value={option.belongTo}
                                        options={selectOptions}
                                        isMulti={true}
                                        //@ts-ignore
                                        onChange={(value) => { changeInputValue(value, index, 'belongTo') }}></Select>
                                </Field>
                            </Card.Description>
                            <Card.SecondaryActions>
                                <IconButton
                                    key="delete"
                                    name="trash-alt"
                                    onClick={() => { removeOptions(index) }}
                                    tooltip="Delete this config" />
                            </Card.SecondaryActions>
                        </Card>
                    </Collapse>
                })
            }
            <Button icon="plus" onClick={() => { addLine() }}>
                Add Options
            </Button>
        </Collapse>
    </>
};

