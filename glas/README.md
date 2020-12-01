﻿# glas - command line tools for preprocessing LAS files

(с) softlandia@gmail.com / sergienko vitaly  
v 0.0.3 4/03/2019

утилита предназначена для предварительной обработки LAS файлов  
программа бесплатна, доступна на (https://github.com/softlandia/glas) 
вы можете применяеть утилиту на свой страх и риск

## Функции

1. конвертация кодовой страницы из CP866 в WIN1251
2. сбор сведений о содержащихся в LAS файлах мнемониках каротажей
3. конвертация LAS файлов с переименованием мнемоник по словарю


## Использование

>glas w "d:\данные скважин\las" "d:\output\обработанные las"

где

- w команда
- "d:\данные скважин\las" - каталог с входными файлами. все файлы останутся без изменения. если путь длинный или в пути имеются пробелы, то кавычки обязательны
- "d:\output\обработанные las" - каталог с выходными файлами. если путь длинный или в пути имеются пробелы, то кавычки обязательны

команды:  
i - сбор информации о файлах, формируются отчёты  
x - конвертация LAS файлов с переименованием мнемоник по словарю


## Особенности

1.  При конвертации *результирующий* файл записывается в кодировке *CodePage Win 1251*
2.  Чтение и запись с заменой в один и тотже каталог не предусматривается, и категорически *не рекомендуется*
3.  Отчёты которые формируются при работе программы - перезаписываются
4.  В каталоге программы содержатся файлы ini, их наличие обязательно для работоспособности
5.  Все сообщения программа выводит на английском языке (в меру его знания автором). Русификация планируется
6.  Обрабатываются *ВСЕ* найденные las фалы в указанном входном каталоге, поиск осуществляется рекурсивно включая подкаталоги
7.  При работе программы формируются отчёты: текстовые файлы с расширением "md" - обычный текстовый файл, но например Far и Notepad++ предлагают 
немного сахара по синтаксической подсветке данных файлов
8.  glas.ini - файл содержит настройки программы (менять можно, но стоит подумать)
9.  mnemonic.ini - файл содержит СТАНДАРТНЫЕ с моей точки зрения мнемоники (такие, какими они нужны мне в итоге). 
10. dic.ini - словать подстановок прочитанной мнемоники на стандартную. перед = стоит мнемоника которую надо заменить, после = на что надо поменять
11. Что делать если в исходных файлах будет мнемоника с символом '=' я пока не думал... подумаю об этом позже :-)
12. Файлы с переносами данных пока не читаются, игнорируются сразу. Конечно я этим займусь.
13. При конвертации данные записываются с тремя знаками после запятой. *Будте осторожны*

## Команда i - info

пример: >glas i "d:\данные скважин\las"

Предназначена для сбора информации каротажах хранящихся в las файлах

Происходит формирование 4-х отчётов 

1. las.message.md	- будет содержать список las файлов которые ВООБЩЕ не удалось обработать и короткое сообщение о причине
2. las.warning.md	- будет содержать список las файлов при чтении которых выявлены различные ошибки, но их удалось игнорировать
3. log.info.md    	- будет содержать список всех las файлов с результатом сапоставления имеющихся мнемоник со словарными
4. log.missing.md 	- будет содержать список всех мнемоник которым не удалось найти словарную подстановку

Файлы которые будут упомянуты в *las.message.md* безусловно необходимо исправить, исключая файлы с параметром WRAP, я полагаю прочитать такие las не сможет ни одна программа
Файл *las.warning.md* просмотрите ВНИМАТЕЛЬНО. Большинство сообщений в нём будут по делу.
Каждая строка содержит одно сообщение. В ней будет указана строка las файла в которой содержится что-то ПЛОХОЕ. Номер строки указан в 4 колонке после "l:" (колонки разделены символом ";")

Причины приводящие к warning:

1. пустая строка данных, но дальше данные есть
2. неверное значение параметра STEP - встречается довольно часто, просто игнорируем, ведь в данных шаг по факту присутствует 
3. ошибки конвертации данных в число - например "6.2.1". Как и кто генерирует такие las файлы я не знаю и даже представить не могу... Заменяется на NULL
4. не регулярный шаг в данных
5. пропуск данных для одного из каротажей. Это если в строке с данными количество кололонок меньше чем каротажей


## Команда x - eXport

пример >glas x "c:\данные по скважинам\" "f:\обработанные файлы\las"

Предназначена для чтения las файлов из "c:\данные по скважинам" и записи в НОВЫЙ каталог
Что получится если писать в каталог из которого читается - предсказать трудно. Но точно получится не то, что должно...

1. las.message.md	- будет содержать список las файлов которые ВООБЩЕ не удалось обработать
2. las.warning.md	- будет содержать список las файлов при чтении которых выявлены различные ошибки, но их удалось игнорировать

## LICENSE
НЕТ НИКАКИХ ГАРАНТИЙ ДЛЯ ПРОГРАММЫ ДО РАМОК, ДОПУСТИМЫХ ДЕЙСТВУЮЩИМ ЗАКОНОДАТЕЛЬСТВОМ. ЕСЛИ ИНОЕ НЕ УСТАНОВЛЕНО В ПИСЬМЕННОЙ ФОРМЕ, 
ПРАВООБЛАДАТЕЛЬ И/ИЛИ ДРУГИЕ СТОРОНЫ ПРЕДОСТАВЛЯЮТ ПРОГРАММУ «КАК ЕСТЬ», БЕЗ КАКИХ ЛИБО ГАРАНТИЙ (ЗАЯВЛЕННЫХ ИЛИ ПОДРАЗУМЕВАЕМЫХ), 
ВКЛЮЧАЯ, НО, НЕ ОГРАНИЧИВАЯСЬ, ПОДРАЗУМЕВАЕМЫМИ ГАРАНТИЯМИ ТОВАРНОГО СОСТОЯНИЯ ПРИ ПРОДАЖЕ И ГОДНОСТИ ДЛЯ ОПРЕДЕЛЕННОГО ПРИМЕНЕНИЯ. 
ВЕСЬ РИСК, КАК В ОТНОШЕНИИ КАЧЕСТВА, ТАК И ПРОИЗВОДИТЕЛЬНОСТИ ПРОГРАММЫ ВЫ БЕРЕТЕ НА СЕБЯ. ЕСЛИ В ПРОГРАММЕ ОБНАРУЖЕН ДЕФЕКТ, ВЫ 
БЕРЕТЕ НА СЕБЯ СТОИМОСТЬ НЕОБХОДИМОГО ОБСЛУЖИВАНИЯ, ПОЧИНКИ ИЛИ ИСПРАВЛЕНИЯ.

НИ В КОЕМ СЛУЧАЕ, ЕСЛИ НЕ ТРЕБУЕТСЯ ПРИМЕНИМЫМ ЗАКОНОМ ИЛИ ПИСЬМЕННЫМ СОГЛАШЕНИЕМ, НИ ОДИН ИЗ ПРАВООБЛАДАТЕЛЕЙ ИЛИ СТОРОН, ИЗМЕНЯВШИХ 
И/ИЛИ ПЕРЕДАВАВШИХ ПРОГРАММУ, КАК БЫЛО РАЗРЕШЕНО ВЫШЕ, НЕ ОТВЕТСТВЕНЕН ЗА УЩЕРБ, ВКЛЮЧАЯ ОБЩИЙ, КОНКРЕТНЫЙ, СЛУЧАЙНЫЙ ИЛИ ПОСЛЕДОВАВШИЙ 
УЩЕРБ, ВЫТЕКАЮЩИЙ ИЗ ИСПОЛЬЗОВАНИЯ ИЛИ НЕВОЗМОЖНОСТИ ИСПОЛЬЗОВАНИЯ ПРОГРАММЫ (ВКЛЮЧАЯ, НО, НЕ ОГРАНИЧИВАЯСЬ ПОТЕРЕЙ ДАННЫХ ИЛИ НЕВЕРНОЙ 
ОБРАБОТКОЙ ДАННЫХ, ИЛИ ПОТЕРИ, УСТАНОВЛЕННЫЕ ВАМИ ИЛИ ТРЕТЬИМИ ЛИЦАМИ, ИЛИ НЕВОЗМОЖНОСТЬ ПРОГРАММЫ РАБОТАТЬ С ДРУГИМИ ПРОГРАММАМИ), 
ДАЖЕ В СЛУЧАЕ ЕСЛИ ПРАВООБЛАДАТЕЛЬ ЛИБО ДРУГАЯ СТОРОНА БЫЛА ИЗВЕЩЕНА О ВОЗМОЖНОСТИ ТАКОГО УЩЕРБА