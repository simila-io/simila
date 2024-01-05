# Query Language (QL)
Simila accepts different types of requests for selecting index records or running a search for text-phrase. These types of requests may require some filtering, which can be an OR/AND combination of conditions over the tags and properties. To make life easier, we introduced the Query Language (QL), which allows specifying filtering conditions. The QL resembles the WHERE expression syntax in SQL, so it doesn't require special knowledge or learning new concepts. We will try to explain it concisely here so that the reader can apply it immediately.

QL is a very simple language for writing boolean expressions. For example:

```
tag('abc') = tag("def") AND (prefix(path, "/aaa/") OR format = "pdf") 
```

the expression above select objects with tag 'abc' value equals to the tag 'def' value AND with which path starts from '/aaa' or the format value is 'pdf'.


## Arguments
An argument is a value, which can be referenced by one of the following forms:
- constant
- identifier
- function
- list of constants

### Constant values
QL supports two type of constants - string and a number. The string constant is a text in double or single quotes. The numbers are either natural or real numbers. 

String constants examples:
```
''
""
"Hello world"
'Andrew said: "Hello!"'
```

And the number's ones:
```
1
-123
34.234243
34.001E-23
```

### Identifiers
Identfier is a variable, which adressed by name. QL supports the following identifiers:
- `path` - the path to an object. The value of the path may look like `/abc/aaa/`
- `node` - the fully-qualified name of a node, it is actualy its path + name. For example `/abc/aaa/doc.txt`
- `format` - the node format. The value may be "pdf", for example.

### Functions
A function is a value that is calculated from the arguments provided. It looks like an identifier followed by arguments in parentheses. The argument list may be empty.

Simila supports the following functions:
- `tag(<name>)` - returns the tag value for the node. Name could be a string constant or any other argument value
- `prefix(<a>, <b>)` - returns either true or false: the a's argument value has the prefix of the b's value.

### List of constants
Some constants maybe groupped in a list. The List defined like the coma-separted constants in between `[` and `]`:

```
["a", "b", 'c'] 
```
the list of three string constants - "a", "b" and "c"

## Operations
Operation is an action which requires two arguments. All operations return either TRUE or FALSE

Examples:
```
'1234' = '234' // compares two string constants, the result will be FALSE
abc != 123 // compares identifier abc with number 123, the result depends on the abc value
tag("t1") > tag("t2") // compares value of the tag t1 with the value of the tag t2, the result depends on the tags values
tag("t1") IN [1, 2, 3] // the value of t1 is either 1, 2, or 3  
```

QL supports the following operations:

"<", ">", "<=", ">=", "!=", "="

| Operation | Description                                                                                                 |
|-----------|-------------------------------------------------------------------------------------------------------------|
| <         | The left argument is less than the right one                                                                |
| >         | The left argument is greater than the right one                                                             |
| <=        | The left argument is less or equal to the right one                                                         |
| >=        | The left argument is greater or equal to the right one                                                      |
| !=        | The left argument is not equal to the right one                                                             |
| =         | The left argument is equal to the right one                                                                 |
| IN        | The left argument value is in the list. Right argument must be a list                                       |
| LIKE      | The left argument should be like the constant (second argument). The operation is similart to the SQL like. |

## QL boolen expression
The QL expression is the series of boolean values that can be combined by AND, OR, NOT boolean operations and the parenthesis to increase the priority.

Please, note, that the `prefix()` function returns a boolean value, so it can be used like a stand-alone argument.

Examples:
```
tag('t1') != tag('t2') and node = '/aaa/doc.txt'
tag('t1') != tag('t2') and (node = '/aaa/doc.txt' or format = 'pdf')
tag('t1') != tag('t2') and prefix(path, '/aaa/')
```

## That is it
With all the information above you can define a filter in a form of QL boolean expression.