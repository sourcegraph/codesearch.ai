// Top-level function
function f() {
    /*
        function f
    */
    return 1; // return
}

// Arrow fn
const a = () => 1 + 1;

// Anon fn
const b = function (params) {
    console.log()
}

// Named function in var
const c = function x() {}

var obj = {
    /**
     * Comment
     * 
     * @returns 1
     */
    g: () => {
        return 1;
    },
    // Multi
    // lines

    // single
    // comment
    h: function () {
        return 2;
    },
}

class C {
    /* Getter */
    get field() {
        return 1
    }

    // Setter
    set field(f) {
        f = f

        const x = () => {
            console.log("nested")
        }
    }

    // Class
    // method
    method() {
        const things = arr.map(function (t) {
            return t.t
        })
        return "method"
    }

    // ToString
    toString() {
        return "C"
    }
}
