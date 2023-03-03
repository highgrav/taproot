# Server-Side Javascript

Taproot embeds the `goja` Javascript runtime, which allows you to execute arbitrary Javascript, either directly as 
server-side Javascripr (SSJS) or using the JSML markup language (see the JSML docs for more on this). Each SSJS script 
is run in a sandboxed JS runtime