import React, { useEffect, useRef, useState } from 'react';
import { PlusOutlined } from '@ant-design/icons';
import { locationService } from '@grafana/runtime';
import { Flex, Input, Tag, theme, Tooltip, InputRef, Space, Select, TourProps, Tour } from 'antd';
import { SimpleOptions } from 'types';

const tagInputStyle: React.CSSProperties = {
    width: 150,
    height: 32,
    marginInlineEnd: 8,
    verticalAlign: 'top',
};

const tagSelectStyle: React.CSSProperties = {
    width: 150,
    height: 32,
    verticalAlign: 'top',
};

interface PanelState {
    options: SimpleOptions,
    data: any,
    width: number,
    height: number,
    replaceVariables: Function
}

function isOptionInCurrentConfig(str: string, options: any[], config: string): boolean {
    const result = options.filter(option => option.value === str)
    if (result.length > 0) {
        return result[0].belongTo.filter((item: any) => item.value === config).length > 0
    }
    return false
}

function getFilterList(options: SimpleOptions, replaceVariables: Function, valueConnect: string): string[] {
    const resultList: string[] = []
    options.filterConfig.map(config => {
        if (!replaceVariables(`$${config.filterKey}`)) {
            return
        }
        const variable = replaceVariables(`$${config.filterKey}`).split(config.optionConnectMark)
        variable.map((item: string) => {
            const str = item.replace(config.keyPrefix, '').replace(config.keySuffix, '').replace(config.valuePrefix, '').replace(config.valueSuffix, '')
            const key = str.split(config.valueConnectMark)[0]
            const value = str.split(config.valueConnectMark)[1]
            resultList.push(key + valueConnect + value)
        })
    })
    return resultList
}

const FilterPanel: React.FC<PanelState> = ({ options, data, width, height, replaceVariables }) => {
    const valueConnect = '='
    const refTag = useRef(null)
    const { token } = theme.useToken();
    const [tags, setTags] = useState<string[]>(getFilterList(options, replaceVariables, valueConnect));
    const [inputVisible, setInputVisible] = useState(false);
    const [inputValue, setInputValue] = useState('');
    const [inputPrefix, setInputPrefix] = useState('');
    const [editInputIndex, setEditInputIndex] = useState(-1);
    const [editInputValue, setEditInputValue] = useState('');
    const [editInputPrefix, setEditInputPrefix] = useState('');
    const [open, setOpen] = useState(false)
    const inputRef = useRef<InputRef>(null);
    const editInputRef = useRef<InputRef>(null);
    const steps: TourProps['steps'] = [
        {
            title: '输入查询条件',
            description: '请点击此处添加查询选项，选择查询条件并输入查询内容。',
            target: () => refTag.current,
        },
    ];

    // 检查localStorage中是否已经存储了标志
    if (!localStorage.getItem('hasSeenIntro')) {
        // 如果没有这个标志，则假设这是用户第一次访问
        // 在这里启动你的漫游式引导
        setOpen(true)
        // 在启动引导后，设置一个标志表示用户已经看过引导
        localStorage.setItem('hasSeenIntro', 'true');
    }
    useEffect(() => {
        if (inputVisible) {
            inputRef.current?.focus();
        }
    }, [inputVisible]);

    useEffect(() => {
        editInputRef.current?.focus();
    }, [editInputValue]);

    const setReplaceVariables = (tags: string[]) => {
        options.filterConfig.map(config => {
            const newList: string[] = []
            tags.map((tag: string) => {
                const prefix = tag.split(valueConnect)[0]
                const value = tag.split(valueConnect)[1]
                if (isOptionInCurrentConfig(prefix, options.options, config.filterKey)) {
                    newList.push(config.keyPrefix + prefix + config.keySuffix + config.valueConnectMark + config.valuePrefix + value + config.valueSuffix)
                }
            })
            locationService.partial({ [`var-${config.filterKey}`]: newList.join(config.optionConnectMark) }, true)
        })
    }

    const handleClose = (removedTag: string) => {
        const newTags = tags.filter((tag) => tag !== removedTag);
        setReplaceVariables(newTags)
        setTags(newTags);
    };

    const showInput = () => {
        setInputVisible(true);
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(e.target.value);
    };

    const handleInputPrefixChange = (value: string) => {
        setInputPrefix(value);
    };

    const handleInputConfirm = () => {
        if (inputValue && !tags.includes(inputValue)) {
            setTags([...tags, inputPrefix + valueConnect + inputValue]);
            setReplaceVariables([...tags, inputPrefix + valueConnect + inputValue])
        }
        setInputVisible(false);
        setInputValue('');
        setInputPrefix('')
    };

    const handleEditInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setEditInputValue(e.target.value);
    };

    const handleEditInputPrefixChange = (value: string) => {
        setEditInputPrefix(value);
    };

    const filterOption = (input: string, option?: { label: string; value: string }) =>
        (option?.label ?? '').toLowerCase().includes(input.toLowerCase());

    const changeTagKey = (tag: string): string => {
        if (tag) {
            const tagKey = tag.split(valueConnect)[0]
            const tagValue = tag.split(valueConnect)[1]
            const tagKetStr = options.options.filter(option => option.value === tagKey)
            if (tagKetStr.length > 0) {
                return tagKetStr[0].label + valueConnect + tagValue
            }
        }
        return '';
    }

    const handleEditInputConfirm = () => {
        const newTags = [...tags];
        newTags[editInputIndex] = editInputPrefix + valueConnect + editInputValue;
        setReplaceVariables(newTags)
        setTags(newTags);
        setEditInputIndex(-1);
        setEditInputValue('');
    };

    const tagPlusStyle: React.CSSProperties = {
        height: 32,
        background: token.colorBgContainer,
        borderStyle: 'dashed',
        fontSize: 14,
        paddingTop: 4
    };

    return (
        <Flex gap="4px 0" wrap="wrap">
            {tags.map<React.ReactNode>((tag, index) => {
                if (editInputIndex === index) {
                    return (
                        <Space.Compact key={'key' + tag}>
                            <Select
                                showSearch
                                defaultValue={''}
                                style={tagSelectStyle}
                                value={editInputPrefix}
                                onChange={handleEditInputPrefixChange}
                                onFocus={() => { setEditInputIndex(index); setEditInputValue(editInputValue) }}
                                filterOption={filterOption}
                                options={options.options} />
                            <Input
                                ref={editInputRef}
                                key={tag}
                                size="middle"
                                style={tagInputStyle}
                                value={editInputValue}
                                onChange={handleEditInputChange}
                                onBlur={handleEditInputConfirm}
                                onPressEnter={handleEditInputConfirm} />
                        </Space.Compact>
                    );
                }
                const isLongTag = changeTagKey(tag).length > 20;
                const tagElem = (
                    <Tag
                        key={tag}
                        closable={true}
                        style={{ userSelect: 'none', fontSize: 14, paddingTop: 4 }}
                        onClose={() => handleClose(tag)}
                    >
                        <span
                            onDoubleClick={(e) => {
                                setEditInputIndex(index);
                                setEditInputPrefix(tag.split(valueConnect)[0]);
                                setEditInputValue(tag.split(valueConnect)[1]);
                                e.preventDefault();
                            }}
                        >
                            {isLongTag ? `${changeTagKey(tag).slice(0, 20)}...` : changeTagKey(tag)}
                        </span>
                    </Tag>
                );
                return isLongTag ? (
                    <Tooltip title={changeTagKey(tag)} key={tag}>
                        {tagElem}
                    </Tooltip>
                ) : (
                    tagElem
                );
            })}
            {inputVisible ? (
                <Space.Compact>
                    <Select
                        showSearch
                        defaultValue={''}
                        value={inputPrefix}
                        style={tagSelectStyle}
                        onFocus={showInput}
                        onChange={handleInputPrefixChange}
                        filterOption={filterOption}
                        options={options.options} />
                    <Input ref={inputRef}
                        type="text"
                        size="middle"
                        style={tagInputStyle}
                        value={inputValue}
                        onChange={handleInputChange}
                        onBlur={handleInputConfirm}
                        onPressEnter={handleInputConfirm} />
                </Space.Compact>
            ) : (
                <Tag style={tagPlusStyle} icon={<PlusOutlined />} onClick={showInput} ref={refTag}>
                    添加查询条件
                </Tag>
            )}
            <Tour open={open} onClose={() => setOpen(false)} steps={steps} />
        </Flex>
    );
};

export default FilterPanel;
