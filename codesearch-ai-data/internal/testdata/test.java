// Class
class J {
    /**
     * Constructor J
     */
    public J() {
        this.x = y;
    }

    int x = 0;

    /* Random */

    // A
    public static void a() {}

    // B
    // C
    @Overrides
    public int b() {
        // Returns 1
        return 1;
    }

    public boolean equals() {
        return false;
    }

    /**
     * Return
     * 1
     */
    @OverridesA
    @OverridesB
    public int b() {
        return 1; // Also returns one
    }
}
