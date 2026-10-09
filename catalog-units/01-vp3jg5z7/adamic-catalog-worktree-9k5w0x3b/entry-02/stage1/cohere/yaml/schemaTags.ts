// Fixed default tag tests from Go cohere and yaml 2.9.0.
import { SchemaTag } from './schemaTag.ts';
export function schemaTags(version: string): SchemaTag[] {
    const tags: SchemaTag[] = [];
    if(version === '1.1') {
        tags.push(new SchemaTag('tag:yaml.org,2002:null', '', 'nullTag', false, '^(?:~|[Nn]ull|NULL)?$'));
        tags.push(
            new SchemaTag('tag:yaml.org,2002:bool', '', 'trueTag', false, '^(?:Y|y|[Yy]es|YES|[Tt]rue|TRUE|[Oo]n|ON)$'),
        );
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:bool',
                '',
                'falseTag',
                false,
                '^(?:N|n|[Nn]o|NO|[Ff]alse|FALSE|[Oo]ff|OFF)$',
            ),
        );
        tags.push(new SchemaTag('tag:yaml.org,2002:int', 'BIN', 'yaml11IntBinTag', false, '^[-+]?0b[0-1_]+$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', 'OCT', 'yaml11IntOctTag', false, '^[-+]?0[0-7_]+$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', '', 'yaml11IntTag', false, '^[-+]?[0-9][0-9_]*$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', 'HEX', 'yaml11IntHexTag', false, '^[-+]?0x[0-9a-fA-F_]+$'));
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:float',
                '',
                'yaml11FloatNaNTag',
                false,
                '^(?:[-+]?\\.(?:inf|Inf|INF)|\\.nan|\\.NaN|\\.NAN)$',
            ),
        );
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:float',
                'EXP',
                'yaml11FloatExpTag',
                false,
                '^[-+]?(?:[0-9][0-9_]*)?(?:\\.[0-9_]*)?[eE][-+]?[0-9]+$',
            ),
        );
        tags.push(
            new SchemaTag('tag:yaml.org,2002:float', '', 'yaml11FloatTag', false, '^[-+]?(?:[0-9][0-9_]*)?\\.[0-9_]*$'),
        );
        tags.push(new SchemaTag('tag:yaml.org,2002:merge', '', 'mergeTag', true, '^<<$'));
        tags.push(
            new SchemaTag('tag:yaml.org,2002:int', 'TIME', 'intTimeTag', false, '^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+$'),
        );
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:float',
                'TIME',
                'floatTimeTag',
                false,
                '^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\\.[0-9_]*$',
            ),
        );
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:timestamp',
                '',
                'timestampTag',
                false,
                '^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})(?:(?:t|T|[ \\t]+)([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2}(\\.[0-9]+)?)(?:[ \\t]*(Z|[-+][012]?[0-9](?::[0-9]{2})?))?)?$',
            ),
        );
    }
    else {
        tags.push(new SchemaTag('tag:yaml.org,2002:null', '', 'nullTag', false, '^(?:~|[Nn]ull|NULL)?$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:bool', '', 'boolTag', false, '^(?:[Tt]rue|TRUE|[Ff]alse|FALSE)$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', 'OCT', 'coreIntOctTag', false, '^0o[0-7]+$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', '', 'coreIntTag', false, '^[-+]?[0-9]+$'));
        tags.push(new SchemaTag('tag:yaml.org,2002:int', 'HEX', 'coreIntHexTag', false, '^0x[0-9a-fA-F]+$'));
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:float',
                '',
                'coreFloatNaNTag',
                false,
                '^(?:[-+]?\\.(?:inf|Inf|INF)|\\.nan|\\.NaN|\\.NAN)$',
            ),
        );
        tags.push(
            new SchemaTag(
                'tag:yaml.org,2002:float',
                'EXP',
                'coreFloatExpTag',
                false,
                '^[-+]?(?:\\.[0-9]+|[0-9]+(?:\\.[0-9]*)?)[eE][-+]?[0-9]+$',
            ),
        );
        tags.push(
            new SchemaTag('tag:yaml.org,2002:float', '', 'coreFloatTag', false, '^[-+]?(?:\\.[0-9]+|[0-9]+\\.[0-9]*)$'),
        );
        tags.push(new SchemaTag('tag:yaml.org,2002:merge', '', 'mergeTag', true, '^<<$'));
    }
    return tags;
}
