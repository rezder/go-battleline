Just a few points on what I learned on programming with go.

* Compile loop
* naming conflict package variable and struct.
* Using function metods and marked as updated. I think change that to
not using named result variables
* package vs file 
* Capital letter used for public variable,types, metods and function.
TODO expand.

# Type 

Go define 3 type of types named type, unamed type and alias.

## Alias

Alias is just byte and rune that are alias for int32 and uint8.
alias can be used almost interchangable.

## Unnamed type

Unnamed type is constructed types like slice []int.

## Named type

Named is language define type like int or string and the
type created in the program like "type myType int", "type myType struct{}",
"type mySlice []int" or type myType myStructType.

### Creating type 

When you name a named type you create a new type the only thing you can do
is cast to the original type and **only if** it is not a pointer. The loosest
of connections.

When you name a unamed type you are just naming a type (my words) and you do not
need to cast to the original type.

Casting of type is not for OOP inheritance as you can not cast pointer.
Anonymous fields in structs is more like inheritance.

# Anonymous fields in structs

Anonymous fields in structs is almost like inheritance made explicit.
The question is should the field be a pointer or not. The answer is the same
for fields with structs it depend on if you want to be able to have another pointer
to the struct because you allready have the Object.field.

