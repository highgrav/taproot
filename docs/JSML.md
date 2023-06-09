# JSML: JavaScript Markup Language

JSML is a small, JSX-like templating language that mixes HTML and Javascript, and compiles to server-side JS.

You can think of JSML as a kind of "reverse JSX" -- rather than being JS that allows embedded HTML, it's HTML that 
embeds server-side Javascript. This makes it substantially similar to traditional "gen one" server-side languages, like 
Coldfusion, JSP, or PHP. While it's not ideal for complex web applications, JSML pages allow you to quickly iterate on 
simple pages that can be mixed with SPA or hybrid SPA applications served out of Taproot.

JSML is exceptionally well-suited for HTMX and similar HTML-fragment-rendering front-end approaches; by using JSML, you 
can trivially render HTML fragments in response to HTMX requests.

JSML is surprisingly performant; rendering a basic HTML page takes about 2ms on a low-powered dev machine, about as much 
time as it takes to serve a very small static file. However, marshaling large amounts of data back and forth, particularly 
when having to do JSON conversions, will have a significant impact on performance.

## Tags
JSML files are compiled to Javascript, so the JSML transpiler is mostly an elaborate exercise in outputting text, with 
some additional functional tags that execute logic or act as syntactic sugar. There are two kinds of tags in JSML, 
*semantic* tags (which look like `<go/>` or `<go.*/>` and have an effect on execution) and *non-semantic* tags (any XML or HTML-style tag that doesn't change 
execution).

The three core semantic tags are `<go/>`, `<go.out/>`, and `<go.val/>`. `<go/>` delineates a server-side JS code block; 
anything within `<go> ... </go>` will be treated as executable code. 

`<go.out>...</go.out>` reverses this: anything 
with a `<go.out/>` block will be treated as raw output. JSML begins in raw output mode, so anything outside of a 
semantic tag block -- even if you don't have any semantic tags whatsoever -- will be treated as raw HTML. This means that 
you only need to use `<go.out/>` when inside a `<go/>` block.

One proviso to the above is that *non-semantic tags will always be output as HTML*, even if they appear within a `<go/>` 
block without a `<go.out/>` block. This is because the parser grabs everything that looks like an XML element, so it can 
determine at compilation time whether it's looking at a semantic or non-semantic element. Thus, the following scripts are 
equivalent:

~~~
<go>
    <go.out>
        <ul>
            <li>Hello, World!</li>
        </ul>
    </go.out>
</go>
~~~
and
~~~
<go>
    <ul>
        <li><go.out>Hello, World!</go.out></li>
    </ul>
</go>
~~~

Working with bare non-semantic tags is generally more convenient than wrapping everything with extraneous `<go.out/>` 
tags.

You can wrap JS expressions in `<go.val>...</go.val>` to have them output the expression; this works in both `<go/>` and 
`<go.out/>` blocks. It's important that, when using `<go.val/>`, you ensure that anything appearing between the tags can 
be evaluated as a single expression:

~~~
<go>
    var x = 42;
</go>
The answer is: <go.val>x</go.val>
~~~

If you're inside a tag, you can use `@`-values to insert a variable as the value of an HTML annotation:
~~~
<go>
let className = "default-class";
if(someTest()){
    className = "other-class";
}
<div class=@className>
</go>
~~~
`@`-values are always output inside double quotes, so the above translates to `<div class="default-class">`.

These three tags are all the markup you need to execute JSML. However, there are other semantic tags that provide 
additional capabilities:

- `<go.include src="..." />`: Inserts another JSML file at the point where the tag appears.


BUGS AND TODOS:
-