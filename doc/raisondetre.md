<!--  -*- Mode: GFM; -*-  -->
<!--
  -- (c) 2015, Jamie A. Jennings
  --
-->

# Why Rosie Pattern Language?

*Disclaimer:* In these notes, as in other posted material, I speak for myself, and not on behalf of IBM.

Analytics depends on good data, but also on *tidy data*, i.e. data that is structured, annotated, normalized, and correlated.  It is generally estimated that 80% of data analysis effort is spent pre-processing the data, i.e. to "tidy it up" [1].  Virtually all systems designed to handle unstructured and semi-structured data (e.g. log files) are based on heavy use of regular expressions.
 
Yet, regular expressions are limited in what kinds of patterns they can recognize, and they are notoriously difficult to understand and maintain [2].  Data tidying systems that depend on collections of regular expressions are fragile.  A big data analytics solution that depends on pre-processing with regular expressions is subject to the high cost of getting the regular expressions right in the first place, and then extending and maintaining the collection of expressions.
 
There are also run-time issues with regular expressions.  "Regex" implementations have highly variable performance characteristics that depend on the input data.  Most will consume exponential time on some input, taking seconds instead of microseconds to process a single line [3].
 
Parsing Expression Grammars (PEGs) do not have these limitations [4].  PEGs have many similarities to regular expressions, and use many of the same operators (like * and +), and features (like character classes, e.g. [a-z] and [:punct:]).  PEGs are more powerful than regular expressions, though, and can recognize recursive structures like XML and JSON.  PEGs also permit linear-time implementations (in the size of the input).
 
Tools for using PEGs have been scarce.  Rosie is a tool for tidying up data that is easy to use, powerful, and efficient.  The Rosie Pattern Language (RPL) is a pattern programming language providing:

*	A readable pattern syntax which can contain whitespace and comments;
*	Pattern composition, for building patterns out of other patterns; 
*	A pattern macro facility for generating new patterns;
*	Packages which separate patterns into different namespaces, making it easy to share pattern definitions;
*	An interactive debugger that can step through pattern matching against sample input;
*	A translator that can convert many existing “regex” into RPL;
*	A compact, efficient run-time for data tidying.
 
Rosie leverages a compact and efficient PEG implementation written in C [5] that includes an interpreter for the Lua language [6].  The RPL compiler and related tools are written in Lua, and the RPL extension language (in which users can add features) is Lua.
 
Rosie and RPL are being used in IBM in several projects to parse, annotate, normalize, and correlate raw data prior to use in analytics implemented in Spark.
 
### References
 
[1] Dasu and Johnson, AT&T Labs, “Data Mining and Data Quality”

[2] "In my experience the hardest part is to get the regular expressions for parsing the log files right" says the author of *Grok Constructor*.  (http://grokconstructor.appspot.com)

[3] "Notice that Perl requires over sixty seconds to match a 29-character string" says Russ Cox, researcher at Google, Bell Labs. (https://swtch.com/~rsc/regexp/regexp1.html)

[4] Ford, "Parsing Expression Grammars: A Recognition-Based Syntactic Foundation", POPL 2004

[5] Ierusalimschy, "A Text Pattern-Matching Tool based on Parsing Expression Grammars", Software: Practice and Experience (2008, John Willey and Sons)

[6] Lua (http://www.lua.org)


<!--     ### Notes for future topics                            -->
<!--                                                            -->
<!--     What makes RPL a language?                             -->
<!--                                                            -->
<!--     *   More words, less syntax                            -->
<!--     *   Composability                                      -->
<!--     *   Packages, namespaces                               -->
<!--     *   Aliases                                            -->
<!--     *   Statements that instruct Rosie on what to do       -->

# Overview of Rosie Pattern Language

## Patterns transform unstructured input into structured output

Rosie uses patterns (written in Rosie Pattern Language, or RPL) to transform unstructured or semi-structured text input into structured output.  Currently, the output is in JSON format, although additional output formats are planned.

Many useful patterns are supplied in the Rosie distribution.  You may write additional patterns, or modify the ones supplied.  If there is already a machine-readable description of the input format (e.g. a log format configuration), then it may be possible to automatically generate Rosie patterns from such a description.

A pattern in Rosie Pattern Language may be specified using literal strings and characters and character sets, or it may be built out of other patterns.  The ability to compose patterns is one of the strengths of RPL.

On start-up, Rosie loads a set of files containing Rosie Pattern Language, i.e. pattern definitions.  One of the patterns defined in these files can then be used to match lines of input; this is the match pattern.  When running, Rosie is applying the match pattern to each line of input.

By default, Rosie processes input one line at a time, where the marker that separates lines is a single newline character.  Rosie generates one JSON structure (output) for each line of input.  The output structure contains an entry for each named component of the match pattern.  Every important field within a pattern has a name, and Rosie’s output pairs each name with the corresponding value parsed out of the input line.

Whatever stage comes after Rosie in the data pipeline can then see every important value in each line of input.  For example, we can use a pattern like this one to parse a simple log entry:

``
syslog = datetime_RFC3339 ip_address {process ":"} message
``

The line above defines the pattern named `syslog` to be a datetime, followed by an ip address, a string that is a process followed by a colon (“:”), and a message.  Each of the identifiers used to define `syslog` (`datetime_RFC3339`, etc.) are defined separately.  For example, `process` is defined as:

``
process = { word {"["int"]"}? }
``

... and matches strings like "sshd[16537]" and "syslogd".  For now, note that pattern elements inside braces `{...}` are elements that are glued together with no whitespace separating them, e.g. "[16537]" and not "[ 16537 ]".

Suppose Rosie is asked to match the `syslog` pattern against a raw line of input data like this one:

``
2015-08-23T03:36:25-05:00 10.108.69.93 sshd[16537]: Did not receive identification string from 208.43.117.11
``

The Rosie output will be a structure as follows: (see NOTE below)

``
[{"datetime_RFC3339":["2015-08-23T03:36:25-05:00",
                      {“date_RFC3339":["2015-08-23"]},
                      {"time_RFC3339":["03:36:25-05:00"]}]},
 {"ip_address":["10.108.69.93"]},
 {"process":["sshd[16537]",
             {“word":["sshd"]},
             {"int":["16537"]}]},
 {"NO_ID":["Did not receive identification string from 208.43.117.11",
           {"ip_address":["208.43.117.11"]}]}]}]
``
		   

From this output, we can see that:
*	Rosie collected each piece of the `syslog` pattern and labeled the parts, e.g. `datetime_RFC3339`  and `ip_address`.
*	The definition of the pattern `datetime_RFC3339` appears to have two parts, a date and a time, and those are separated out and labeled as well.
*	The `message` pattern must be an alias for some other patterns, because instead of seeing "message" as a label in the output, we see instead `NO_ID`.
*	The definition of the `NO_ID` pattern includes an ip address, we can infer, because Rosie parsed an `ip_address` out of the message and labeled it.

With structured output such as that above, the raw data has been transformed into something structured, with semantic tags.  It is now possible to work with the data, such as in an analytics process.

**NOTE:**

There are several output formats; JSON is shown.  Also, the entire original line of input (which matches the `syslog` pattern as a whole) can optionally be included in the output, but is omitted along with some other output for clarity.

## The Rosie Pattern Engine is meant to be used early in a pipeline

As a stand-alone executable, Rosie is suitable for processing files containing lines of input.  Rosie may also be used to process streaming input, or "micro-batches".  Rosie’s small executable size and quick start-up time make it possible to invoke Rosie as part of a data processing pipeline.

In addition to simply recognizing items of interest in the input, Rosie can be instructed to transform values, enumerate values, and perform other operations while processing each line.  Normalization is one kind of transformation, e.g. timestamps in various formats may be normalized to an integer number of milliseconds since the epoch.  Sanitizing, such as encryption of sensitive fields, is another kind of transformation.

Enumeration, on the other hand, is a meta-data calculation.  If Rosie is instructed to enumerate a particular field, then all values for that field that appear in the input will be collected into a table.  For example, which host names were seen in the input?  Or, which queue names?  An enumeration table is one piece of meta-data, and is saved so it can be used by subsequent stages in the data processing pipeline.

Another piece of meta-data is the range of values that a particular field has in the given input.  When a field has values that can be compared in the sense of "value x precedes value y", Rosie can be instructed to save the range of a field seen in the input.  When processing a file of log entries, for example, the range of the timestamp field tells you the time period spanned by those entries.

Since Rosie’s extension language is Lua, a full language, real-time processing such as transformations and enumerations can have side effects and access stored state.  As a result, some interesting new uses of Rosie are possible.  One example is to instruct Rosie to parse particular values out of an input stream and take action (send a message, run a script) as the values change in a specified way.

## Rosie Pattern Language is similar to Regular Expressions

Rosie patterns are Parsing Expression Grammars (PEGs), not Regular Expressions (regex), although the two have much in common.  If you know regex, you know most of the important features of PEGs, and the rest are easy to learn.  [See below](#regex_and_rpl) for more.

Rosie's pattern language, RPL, inverts the usual regex concept that characters match themselves except when they are _special_, like "." and "?" and so many more.  In RPL, to match a literal string of characters, you quote them.  Within a quoted string, the only special character is the backslash, which is used as the escape character.

A string of characters outside of quotes, like `word` or `syslog` is an identifier that refers to a previously defined pattern.  In this way, RPL is like a programming language: you can give names to patterns, and build new patterns out of old ones.

Outside of a quoted string, there are familiar operators that have the same meanings as they do in regex, such as ".", "*", "?", etc.  There's never any confusion about when these characters are operators and when they are literal characters to be matched.  If a character appears in a character set `[...]` or a literal string `"..."`, then it's a literal character to be matched.

### Grouping

In regex syntax, parentheses create both a group and a capture, except when they don't.  (There are many rules and many cryptic syntaxes to remember when writing regex.)  In RPL, parentheses `(...)` and curly braces `{...}` are always used for grouping, just like parentheses are used in mathematical and other expressions in most programming languages.  Parentheses group expressions in _normal mode_, and curly braces group expressions in _raw mode_.

In the normal mode, Rosie tokenizes the input, using punctuation and (any amount or kind of) whitespace as token boundaries.  This is kind of like the "word boundary" `\b` concept from regex.  Normally, Rosie behaves as if there is a regex word boundary between each element in a pattern.

In raw mode, Rosie does no tokenizing.  Patterns are matched character by character, as is common for regex matching.

Having both modes makes it easy to switch back and forth between low-level syntax specifications (in raw mode) and higher level patterns composed of other patterns.

We don't worry about captures when writing RPL expressions.  Everything that has a name (an identifier) is captured automatically.  And there are ways to instruct Rosie definition which parts of a pattern to keep and which to discard, _outside of the pattern definition_.  This makes it possible to write a pattern once, and use it in multiple ways, sometimes keeping most or all of the captures (which we call matches) and other times keeping very little.  Sharing patterns and maintaining patterns are both easier when you can declare outside of the pattern definition itself how you want the matches handled.  Read on for more about how RPL is a small language in ways that regex are not.

### RPL is a small language

The Rosie pattern language intentionally resembles a programming language and not the cryptic syntax-laden regex form.  Recall that, in RPL, literal strings must be appear in double quotes, as in most programming languages.  Also, in RPL, whitespace and comments can appear in pattern definitions _without changing the definition_.  This makes RPL very readable when compared to regex.

RPL statements are typically definitions, where a pattern is assigned a name, just like an assignment statement in programming.  There are also aliases, which are very simple today but will grow into a macro facility.  Pattern names and aliases live in set of namespaces that form a package system, which helps organize patterns, prevent name conflicts, and encourage pattern sharing.  Finally, there are statements in RPL that are instructions which tell Rosie how to handle matches.  For any given pattern, you can tell Rosie to keep certain sub-matches (which are like regex captures) and discard others.  You can tell Rosie to transform a match on the fly, e.g. to convert timestamps into a canonical format such as microseconds since the epoch.  You can tell Rosie to collect meta-data, such as the range of values seen.

An RPL expression may not be as compact as its regex equivalent (where such equivalent exists), but RPL is easier to write, understand, use, and maintain.

See [below](#regex_and_rpl) for additional information on the differences between RPL and regexes.

### Why doesn't Rosie use regex?

Rosie’s patterns are not regular expressions, though they share with regular expressions the ability to “recognize a language”.  That is a formal way of saying that a Rosie pattern expression, like a regular expression, can be used to recognize certain input strings and reject others.  Which strings are recognized and which are rejected depends, of course, on the pattern itself.

True regular expressions come from computing theory: they are expressions that recognize exactly the same sets of strings that can be recognized using a simple finite state machine.  Two important properties of regular expressions are:

1. Because they are based on a formal system, many important properties of regular expressions can be proved; and
2. The complexity of recognizing or rejecting an input string is linear in the length of the input.

Property (1) is important for many reasons, one of which is that formal proofs allow implementations to transform regular expressions into equivalent ones that may execute faster.  Another benefit of a formal basis is that it tells us how to compose simple expressions to make more complex ones.  Property (2) reassures us that our software will not consume excessive memory or time when trying to match an expression against an input string.

However...

These properties hold only for regular expressions that stick to the formal definition.  In practice, what we call “regex” libraries (like PCRE, Perl regex,  Java regex) contain significant extensions to the regular expression formalism, such as numbered/named captures, counted repetition, and back references.  With these extensions, it becomes very hard to robustly compose regexes, and the performance of matching using a regex is quite variable. It can take exponential time in the length of the input!

Some other limitations are shared by the formal regular expressions and the
widely used regex tools, such as the inability to match recursively defined
input, such as XML or JSON.  In computing theory, one turns to a context-free
grammar (CFG) to write such patterns.  Programming tools for using a CFG are not commonly
available, though, and consequently most programmers are familiar only with one
matching technology: the regex.

In ordinary use, the regex technology works well.  Small regex patterns embedded in large programs allow easy extraction of data from strings; and regex patterns composed on the command line greatly aid administrative tasks.  But when we have lots of patterns, regex solutions are fragile and unmaintinable.  And when we have large amounts of data, the variable nature of regex performance can impact an entire solution.

### Regex expressions are not maintainable

When regex patterns become long, they become unreadable, even by the person who wrote them.  A collection of non-trivial regexes is not maintainable in the way that, say, Python or Java code is.  In this sense, regex is a poor basis for anything but the most trivial pattern matching tasks, because the regex approach does not scale as the patterns become more numerous and more complex.

Some examples comparing regex-based tools like Grok to Rosie are in the table below.  Note the similarity of syntax where the concept is the same, such as the `*`, `+`, and `?` to specify repetition.  And note how the relative lack of special characters in RPL means that the list of allowed characters in a Unix path can be written simply as `[_%!$@:.,~-]`.  All of the characters inside the square brackets simply represent themselves.  None are special operators.  (The lone exception in character sets is the dash that separates two characters in a range such as `[a-z]`.)

<!-- Vertical bars cannot be escaped in markdown, so we have to use the -->
<!-- character syntax &#124; which does not work in backticks, hence the need -->
<!-- for the html <code/> tags.  And &#42; for the star. And &#94; for the caret.-->

| Grok and other regex tools        | Rosie Pattern Language | Comment       |
| --------------------------------- | ---------------------- | --------       |
| `INT = (?:[+-]?(?:[0-9]+))`       | `int = { [+-]? digit+ }` | These are comparable in readability.  In RPL, `digit` is an alias for `[:digit:]`.        |
| <code>PATH (?:%{UNIXPATH}&#124;%{WINPATH})</code> |  `path = unix_path / windows_path` | RPL is a bit cleaner, and uses `/` to mean "ordered choice". |
| <code>UNIXPATH (?>/(?>[\w_%!$@:.,~-]+&#124;\\.)&#42;)+</code> | `unix_path = {"/" {[:alnum:]/[_%!$@:.,~-]}+ }+`| What is `(?>` in regex? Should look that up. |
| <code>WINPATH (?>[A-Za-z]+:&#124;\\)(?:\\[&#94;\\?&#42;]&#42;)+</code> | `windows_path = { {[:alpha:]+ ":"}? {"\\" {![\\?*] any}* }+ }` | The RPL `!` means "not looking at" |
| <code>/&#94;\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{2,}</code><br><code>\\+\\d{4}[\\s]+\\[([\\w]+)\\/([\\d]+)\\][\\s]+(OUT&#124;ERR)</code><br><code>[\\s]+.&#42;\\[(\\d{4}-\\d{2}-\\d{2}\\d{2}:\\d{2}:\\d{2}\\.\\d{2,})\\]</code><br><code>[\\s]+\\[(.&#42;)\\][\\s]+(.&#42;)[\\s]+-[\\s]+.&#42;\\[\\d{2}m(.&#42;)\\]</code> | | This regex was written by an expert programmer.  Any regex of more than about 12 characters becomes impenetrable due to the dense syntax. |

### Regex solutions may not scale to large data

The other scale issue we must consider is the size of the input (data).  When data sets get large, the unpredictable nature of regex processing time becomes an issue.  Typical input may require fractions of a millisecond to process, but what happens to the overall data pipeline when some input arrives that bogs down the regex implementation, requiring several seconds per line?  The pipeline stalls horrendously.

In ordinary use on the command line (such as listing files with `ls` or searching in files with `grep`), regex performance is not an issue.  When a regex is used to extract some information from a string in Java, performance is not usually an issue.  But when regex patterns are deployed to parse large amounts of data, two things can go wrong: backtracking done by regex engines can destroy the performance, and malformed input can force so much backtracking that exponential time is required.

Since regexes are not maintainable and fail to scale in both pattern size and input size, Rosie is based instead on a different pattern expression approach: Parse Expression Grammars.


## Parse Expression Grammars

As we mentioned earlier, the Context Free Grammar is more powerful than the regular expression.  With that power comes inefficient matching (e.g. the widely cited CYK algorithm has an upper bound of _n<sup>3</sup>_, where _n_ is the input length) and a cumbersome way to specify a pattern (always writing a full grammar).

Parse Expression Grammars were defined to avoid the pitfalls of CFGs, at the expense of being less powerful.  It turns out that PEGs are more powerful than regular expressions (and regex), though, and may be seen as sitting somewhere in between regex and CFGs.

A PEG requires only linear time (in the length of the input) to process an input string.  It does this by limiting backtracking: the "alternation" operator in a PEG is an _ordered choice_.  Instead of using the pipe symbol `|`, which represents equal alternatives in regex, a PEG uses a forward slash `/` to denote an ordered choice between two alternatives.  The PEG pattern `(a / b) c` is read as "a or b, followed by c" and is processed this way:

1. If `a` matches the start of the input, then the choice is satisfied, so go on to match `c`
2. Else if `b` matches the start of the input, then the choice is satisfied, so go on to match `c`
3. Otherwise, the entire pattern fails because `(a / b)` could not be matched

The difference between the PEG choice operator `/` and the regular expression `|` becomes clear with a different example.  The PEG pattern `a / (a b)` will not match the input "a b" because the pattern will never look for b, given that input!  This pattern is processed as follows:

1. If `a` matches the start of the input, then the ordered choice is satisfied, and the overall pattern succeeds
2. Else try the next alternative, `(a b)`.  In this example, `(a b)` will always fail because we arrived here due to the fact that we could not match `a`.  If we cannot match `a`, then we cannot match the sequence `(a b)`.

The order of the alternatives matters in PEG.  The pattern `(a b) / a` will match input "a b", because this pattern looks for the sequence "a b" first.

When writing PEG patterns, then, we must pay attention to the order of choices in an alternation expression.  Ordered choices and sequences are just part of Parse Expression Grammars.  The full capabilities of PEGs are as follows (note the similarity to regexes):

|  PEG expression | Meaning                      |
|  -------------- | -------                      |
|  `pat?`         | Zero or one instances of `pat`                      |
|  `pat*`         | Zero or more instances of `pat`                      |
|  `pat+`         | One or more instances of `pat`                      |
|  `pat?`         | Zero or one instances of `pat`                      |
|  `!pat`         | Not looking at `pat` (consumes no input)                      |
|  `@pat`         | Looking at `pat` (consumes no input)                       |
|  `p / q`        | Ordered choice between `p` and `q`     |
|  `p q`          | Sequence of `p` followed by `q`     |

**NOTE:**  The "quantified expressions" `pat?`, `pat*`, `pat+`, and `pat?` are greedy versions of those used in regular expressions.  They will consume as many repetitions as possible, always.

The Rosie Pattern Language adds some additional features to the PEG formalism:

|  RPL expression | Meaning                      |
|  -------------- | -------                      |
|  `(...)`         | Grouping, as in mathematics, to force order of operations |
|  `{...}`         | Raw group, which tells Rosie to process character by character |
|  `pat{n,m}`         | Bounded repetition of `pat`.  Each of `n` and `m` are optional. |


## Captures in Rosie Pattern Language

### Patterns are assigned names

Any pattern matching technology is made more useful with the concept of a capture, which simply refers to the portion of the input that matched a pattern.  When writing RPL patterns, we assume there are certain portions of the input you want to single out and label, e.g. the command name and the URL in an HTTP command.

Since RPL patterns can be made out of other patterns, you can write a pattern that matches an entire line of input by using patterns that have already been defined.  For example,

``` 
http_command = http_command_name (url / path)
```


is a statement in RPL that defines a pattern named `http_command` as an `http_command_name` followed by either a `url` or a `path` – i.e. the definition refers to three other patterns.

When Rosie matches the pattern `http_command`, the resulting output is a data structure that reflects the structure of the match.  By default, the output includes the portion of the input that matched, e.g. "GET http://ibm.com/index.html", as well as the matches for `http_command` and (in this case) `url`.  Such output, in JSON, might look like this:

``` 
{"http_command": ["GET http://ibm.com/index.html",
                   {"http_command_name":["GET"]},
                   {"url":["http://ibm.com/index.html"]}]}
``` 
				   
This is default behavior, and it allows the consumer of Rosie's output to see all of the important "fields" (like `url`) and their values (like http://ibm.com/index.html).

Rosie provides additional controls over what is captured and what is discarded.  For instance, you may not have any need for the entire original input (i.e. the string that matched http_command above), so you can tell Rosie to discard it.  And you may want to annotate this particular `http_command` with additional information based on where it occurred in the input. Finally, you may want to transform some matched values, or enumerate the values seen, etc.

RPL provides statements that serve as instructions for Rosie to do all of these things.

**NOTE:** Documentation of these features is forthcoming and will appear here as
  the syntax for those features is finalized.

### Aliases

In RPL, patterns are often built from other patterns.  But sometimes you know that there’s no need to keep all the little parts that make up a match.  An alias in RPL is a pattern definition that is meant to be reused in another pattern, but which does not save the sub-matches of its components.  For example, integers may be defined this way, using two definitions:

``` 
d = [:digit:]
common.int = { [+-]? d+ }
``` 

The curly braces that surround the expression on the right hand side of `common.int` are a _raw group_, which Rosie will process character by character instead of separating the input into tokens first.  The definition of `common.int` reads this way: an optional sign, followed by one or more digits. In this example, `d` is a shorthand to avoid writing `[:digit:]`.  Matching `common.int` against "421" produces:

``` 
{"common.int":["421",{"d":["4"]},{"d":["2"]},{"d":["1"]}]}
``` 

Assuming we don’t need the individual digits of "421" in subsequent processing, we can avoid capturing them at all by defining `d` as an _alias_:

```
alias d = [:digit:]
common.int = { [+-]? d+ }
``` 

Matching against "421" now gives:

``` 
{"common.int":["421"]}
```  

## If you know regex, this is Rosie
<a name="regex_and_rpl"></a>

Forthcoming summary of the differences will go here.

---
*Disclaimer:* In these notes, as in other posted material, I speak for myself, and not on behalf of IBM.
