import React, { useState } from 'react';
import { Select, Button, Input, Card, IconButton, Field, ValuePicker } from '@grafana/ui';
import { SelectableValue, StandardEditorProps } from '@grafana/data';
import './SimpleEditor.css';
import { FieldConfig } from 'types';

interface TypeOption {
    label: string,
    value: string
}

const typeOptions: TypeOption[] = [
    {
        label: 'Text',
        value: 'text'
    },
    {
        label: 'DataLink',
        value: 'dataLink'
    }
]

export const SimpleEditor = ({ item, value, onChange, context }: StandardEditorProps<FieldConfig[]>) => {
    const options: SelectableValue[] = [];
    const [valueCopy, setValueCopy] = useState<FieldConfig[]>(JSON.parse(JSON.stringify(value)))


    const addLine = (name: string) => {
        setValueCopy([...valueCopy, { name: name, type: 'text', overrideName: '', url: '' }])
    }

    const removeLine = (index: number) => {
        valueCopy.splice(index, 1);
        setValueCopy([...valueCopy])
        onChange(valueCopy)
    }

    const changeValue = () => {
        setValueCopy([...valueCopy])
        onChange(valueCopy)
    }

    const updateCopyValue = (index: number, newName: string, newType: string, newOrName: string, newUrl: string) => {
        valueCopy[index].name = newName;
        valueCopy[index].type = newType;
        valueCopy[index].overrideName = newOrName;
        valueCopy[index].url = newUrl;
        setValueCopy([...valueCopy])
    }

    if (context.data) {
        const frames = context.data[0]?.fields;
        for (let i = 0; i < frames?.length; i++) {
            options.push({
                label: frames[i].name,
                value: frames[i].name,
            });
        }
    }
    return <>
        {
            valueCopy.map((line, index) => {
                return <Card key={index}>
                    <Card.Heading>Override Field {index + 1}</Card.Heading>
                    <Card.Description>
                    <Field label='Field Name' description=''>
                            <Select options={options} value={line.name} onChange={(selectableValue) => updateCopyValue(index, selectableValue.value, line.type, line.overrideName, line.url)} />
                        </Field>
                        <Field label='Override Type'>
                            <Select options={typeOptions} value={line.type} onChange={(selectableValue) => updateCopyValue(index, line.name, selectableValue.value as string, line.overrideName, line.url)} />
                        </Field>
                        <Field label='Override Name'>
                            <Input value={line.overrideName} onChange={(e: any) => { updateCopyValue(index, line.name, line.type, e.target.value,line.url) }}></Input>
                        </Field>
                        <Field label='Url' style={{display: line.type === 'dataLink' ? 'flex' : 'none'}} description='You can use the keyword “$param" to obtain panel parameters, or specify your own parameters '>
                            <Input value={line.url} onChange={(e: any) => { updateCopyValue(index, line.name, line.type, line.overrideName, e.target.value) }}></Input>
                        </Field>
                    </Card.Description>
                    <Card.Figure>

                    </Card.Figure>

                    <Card.Actions>
                        <Button key="saves" variant="secondary" onClick={() => { changeValue() }}>
                            Save
                        </Button>
                    </Card.Actions>
                    <Card.SecondaryActions>
                        <IconButton key="delete" name="trash-alt" onClick={() => { removeLine(index) }} tooltip="Delete this data source" />
                    </Card.SecondaryActions>
                </Card>
            })
        }
        <div className='button-ctl'>
            <ValuePicker label='Add' fill="solid" variant="primary" size='md' options={options} onChange={(v) => addLine(v.value)} />
        </div>
    </>
};

