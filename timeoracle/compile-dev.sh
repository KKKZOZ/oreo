#!/bin/bash

go build .
mv timeoracle ../benchmarks/cmd/

times in msec
 clock   self+sourced   self:  sourced script
 clock   elapsed:              other lines

000.022  000.022: --- NVIM STARTING ---
000.263  000.242: event init
000.403  000.140: early init
000.499  000.095: locale set
000.592  000.094: init first window
001.022  000.429: inits 1
001.048  000.026: window checked
001.053  000.005: parsing arguments
002.384  000.546  000.546: require('vim.shared')
003.746  000.565  000.565: require('vim._options')
003.751  001.361  000.796: require('vim._editor')
003.755  002.095  000.189: require('vim._init_packages')
003.759  000.611: init lua interpreter
003.973  000.214: expanding arguments
004.053  000.080: inits 2
004.418  000.365: init highlight
004.422  000.004: waiting for UI
004.619  000.198: done waiting for UI
004.625  000.006: clear screen
004.849  000.224: init default mappings & autocommands
005.814  000.098  000.098: sourcing /usr/share/nvim/runtime/ftplugin.vim
005.912  000.037  000.037: sourcing /usr/share/nvim/runtime/indent.vim
006.714  000.611  000.611: require('user.options')
006.992  000.273  000.273: require('user.keymaps')
007.034  000.991  000.107: sourcing /home/kkkzoz/.config/nvim/init.lua
007.050  001.074: sourcing vimrc file(s)
007.548  000.396  000.396: sourcing /usr/share/nvim/runtime/filetype.lua
007.924  000.154  000.154: sourcing /usr/share/nvim/runtime/syntax/synload.vim
008.231  000.559  000.405: sourcing /usr/share/nvim/runtime/syntax/syntax.vim
009.492  000.224  000.224: sourcing /usr/share/nvim/runtime/plugin/gzip.vim
009.598  000.021  000.021: sourcing /usr/share/nvim/runtime/plugin/health.vim
010.634  000.312  000.312: sourcing /usr/share/nvim/runtime/pack/dist/opt/matchit/plugin/matchit.vim
010.866  001.228  000.915: sourcing /usr/share/nvim/runtime/plugin/matchit.vim
011.095  000.194  000.194: sourcing /usr/share/nvim/runtime/plugin/matchparen.vim
011.684  000.547  000.547: sourcing /usr/share/nvim/runtime/plugin/netrwPlugin.vim
012.000  000.226  000.226: sourcing /usr/share/nvim/runtime/plugin/rplugin.vim
012.198  000.102  000.102: sourcing /usr/share/nvim/runtime/plugin/shada.vim
012.291  000.028  000.028: sourcing /usr/share/nvim/runtime/plugin/spellfile.vim
012.446  000.118  000.118: sourcing /usr/share/nvim/runtime/plugin/tarPlugin.vim
012.629  000.122  000.122: sourcing /usr/share/nvim/runtime/plugin/tohtml.vim
012.695  000.020  000.020: sourcing /usr/share/nvim/runtime/plugin/tutor.vim
012.931  000.187  000.187: sourcing /usr/share/nvim/runtime/plugin/zipPlugin.vim
013.672  000.241  000.241: sourcing /usr/share/nvim/runtime/plugin/editorconfig.lua
013.962  000.245  000.245: sourcing /usr/share/nvim/runtime/plugin/man.lua
014.075  000.078  000.078: sourcing /usr/share/nvim/runtime/plugin/nvim.lua
014.099  002.513: loading rtp plugins
014.422  000.323: loading packages
015.187  000.765: loading after plugins
015.221  000.034: inits 3
019.398  004.176: reading ShaDa
019.834  000.437: opening buffers
019.861  000.026: BufEnter autocommands
019.866  000.005: editing files in windows
019.976  000.110: VimEnter autocommands
019.982  000.006: UIEnter autocommands
804.907  784.849  784.849: sourcing /usr/share/nvim/runtime/autoload/provider/clipboard.vim
804.939  000.108: before starting main loop
805.088  000.149: first screen update
805.092  000.004: --- NVIM STARTED ---


times in msec
 clock   self+sourced   self:  sourced script
 clock   elapsed:              other lines

000.023  000.023: --- NVIM STARTING ---
000.349  000.326: event init
000.508  000.159: early init
000.605  000.097: locale set
000.711  000.106: init first window
001.244  000.533: inits 1
001.260  000.016: window checked
001.266  000.006: parsing arguments
002.610  000.552  000.552: require('vim.shared')
003.821  000.502  000.502: require('vim._options')
003.834  001.217  000.715: require('vim._editor')
003.837  001.953  000.184: require('vim._init_packages')
003.842  000.623: init lua interpreter
005.324  001.482: expanding arguments
005.406  000.082: inits 2
005.785  000.379: init highlight
