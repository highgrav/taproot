
# Goldfusion


Gof is basically JavaScript with some syntactic sugar to let it emit HTML, similar to JSX.

In general, gof files are parsed as HTML, with the text between tags interpreted as Javascript.

- <go.out/> Anything within a <go.out/> tag is treated as text to be emitted, unless overridden within a <go/> tag.
- <go/>: Anything within a <go/> tag is treated as JS.
- <go.var/>: Anything within a <go.var/> tag is treated as a JS expression to output.


#### TODO
- Need to preparse prior to HTML tokenization, since tag-like strings in squots/dquots can be improperly parsed as HTML.

