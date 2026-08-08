import remarkGfm from "remark-gfm";
import remarkMath from "remark-math";
import { remarkMathPolicy } from "./remarkMathPolicy";
import { remarkLocalPathLinks } from "../lib/localPathLinks";

// One shared parser policy keeps live Markdown and session exports identical.
export const inxRemarkPlugins = [remarkGfm, remarkMath, remarkMathPolicy, remarkLocalPathLinks];
export { inxRehypePlugins } from "./rehypeInxKatex";
