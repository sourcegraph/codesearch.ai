def a() -> None:
    "Comment 1"
    1+1

# Above decorator
@decorator
def b(c, d) -> int:
    """
    Comment 2
    Comment 3
    """
    if a:
        b()
    else:
        print(3)
    return c + 2

class E:
    # Method
    def f() -> str:
        # Inner
        def f_nested():
            """
            Comment 4
            """
            print(f"1 {x} 2") # Print
            return f

        1 # 1
        return "abc"
    
    @setter1
    @setter2
    def g():
        "Comment 4"
        pass
