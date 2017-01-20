# html-cruncher

HTML cruncher rewrites IDs, classes and names in HTML, CSS and
JavaScript files in order to save bytes and obfuscate code just like
[html-muncher](https://github.com/ccampbell/html-muncher/) does.

## Example

### Input

HTML:

```HTML
<div id="test1"></div>
<div id="test2"></div>
<div class="test3"></div>
```

Original CSS:

```CSS
#test1, #test2, .test3 {
    background-color: red;
}

.test3 {
    color: white;
}
```

Original JS:

```
function change1(test1) {
    document.getElementById('test2');
    document.getElementById("test2");
    document.getElementsByClassName("test2");
    $('#test1').val('test2');
}

function change2(test2) {
    $(".test2").hasClass('test3');
}

function change3(test3) {
    $('#test3.test2').val('test3');
    $('.test3#test2').val('test3');
}
```

### HTML cruncher output:

HTML:

```HTML
<div id="a"></div>
<div id="c"></div>
<div class="b"></div>
```

CSS:

```CSS
a, c, .b {
    background-color: red;
}

.b {
    color: white;
}
```

JS:

```JS
function change1(test1){
document.getElementById('c');
document.getElementById("c");
document.getElementsByClassName();
$('#a').val('test2');
}
function change2(test2){
$(".test2").hasClass('b');
}
function change3(test3){
$('#test3.test2').val('test3');
$('.b#c').val('test3');
}
```

(Yes, html-cruncher can't restore original formatting of the JavaScript file.)

## Installation

### Installing dependencies

```
go get golang.org/x/net/html
go get github.com/gorilla/css
```

### Installing the app

```
go get github.com/nonoo/html-cruncher
go install github.com/nonoo/html-cruncher
```

## Usage

You can get the available command line options with the **-h** switch.

## TODO

- Rewriting name identifiers
