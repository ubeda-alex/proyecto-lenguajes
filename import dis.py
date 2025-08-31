import dis

# def saludo(nombre):
#     return f"Hola, {nombre}!"

# dis.dis(saludo)


# def suma(a, b):
#     return a + b
# dis.dis(suma)

def contar(lista):
    total = 0
    for n in lista:
        total += n
    return total
dis.dis(contar)