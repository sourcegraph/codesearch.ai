<?php

/**
 * Docstring
 */
function a($b, $c): string {
    // Concat
    return $b + $c + "d";
}

class C {
    // Class comment

    // Method
    // comment
    function f() {}

    private function g() {
        $a = 1 + 1; // Sum up
        return $a;
    }

    // Desctructor
    function __destruct() {}
}
