// Decode checker metadata, validating every field and graph link.
import { panic } from 'adamic';
import { Frames, header } from './frames.ts';
import { TypeFact } from './type_fact.ts';
import { Types } from './types.ts';

function contains(records: readonly TypeFact[], id: number): boolean {
    for(const record of records) {
        if(record.id === id) {
            return true;
        }
    }
    return false;
}

export function types(text: string, question: string): Types {
    const frames = new Frames(text);
    header(frames, question);
    const strict = frames.yes();
    const present = frames.yes();
    const roots = frames.ids();
    const rests: boolean[] = [];
    const restCount = frames.natural();
    for(let index = 0; index < restCount; index++) {
        rests.push(frames.yes());
    }
    const records: TypeFact[] = [];
    const count = frames.natural();
    for(let index = 0; index < count; index++) {
        const id = frames.natural();
        const flags = frames.natural();
        const name = frames.field();
        const error = frames.yes();
        const target = frames.natural();
        const array = frames.yes();
        const tuple = frames.number();
        const constraint = frames.natural();
        const element = frames.natural();
        const parts = frames.ids();
        const arguments_ = frames.ids();
        if(id === 0 || tuple < -1 || contains(records, id)) {
            panic('invalid checker type record');
        }
        records.push(
            new TypeFact(id, flags, name, error, target, array, tuple, constraint, element, parts, arguments_),
        );
    }
    frames.end();
    const result = new Types(strict, present, roots, rests, records);
    for(const root of roots) {
        result.type(root);
    }
    for(const record of records) {
        for(const id of record.parts) {
            result.type(id);
        }
        for(const id of record.arguments) {
            result.type(id);
        }
        if(record.constraint !== 0) {
            result.type(record.constraint);
        }
        if(record.element !== 0) {
            result.type(record.element);
        }
    }
    return result;
}
