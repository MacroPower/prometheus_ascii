# prometheus_ascii
Prometheus client that outputs ASCII graphs

## Usage

```
$ ./prometheus_ascii --prometheus.query="sum(increase(wakatime_seconds_total[1h]))"
```

## Output

```text
  1.2 hr  ┼                               ╭╮                                 
  1.1 hr  ┤                               ││                              ╭╮ 
   56 min ┤                               ││                     ╭──╮ ╭╮  ││ 
   49 min ┤                               ││                     │  │ ││ ╭╯│ 
   42 min ┤                               ││                    ╭╯  │ │╰─╯ ╰ 
   35 min ┤                              ╭╯╰╮                   │   ╰─╯      
   28 min ┤                      ╭╮      │  │                   │            
   21 min ┤                     ╭╯│  ╭─╮ │  │                   │            
   14 min ┤                     │ │  │ ╰╮│  │                  ╭╯            
    7 min ┤                    ╭╯ ╰╮╭╯  ╰╯  │                  │             
    0 ms  ┼────────────────────╯   ╰╯       ╰──────────────────╯             
         sum(increase(wakatime_seconds_total[1h]))
```
