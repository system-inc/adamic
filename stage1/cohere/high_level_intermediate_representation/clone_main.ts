import { panic, programArguments } from 'adamic';
import { constructionCoverage } from './clone_coverage.ts';
constructionCoverage(programArguments()[0] ?? panic('missing manifest'),false,true);
