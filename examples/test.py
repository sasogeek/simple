import time

start_time = time.time()

my_integer = 1
my_float = 1.5
my_string = "Hello Simple!"
my_array = [1, 2, 3]
my_dict = {"name": "Simple", "age": 1}

print(my_integer)
print(my_float)
print(my_string)
print(my_array)
print(my_dict)
print()


if my_integer > my_float:
    print("Integer is bigger")
else:
    print("Float is bigger")
print()


print("Counting from 1 to 10:")
for i in range(10):
    print(i+1)
print()

x = 10
print("Counting down from 10")
while x > 0:
    print(x)
    x = x - 1

######################
elapsed_time = time.time() - start_time
print(f"Elapsed time: {elapsed_time:.8f} seconds")
