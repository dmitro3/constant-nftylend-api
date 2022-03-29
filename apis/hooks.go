package apis

import (
	"net/http"

	"github.com/czConstant/constant-nftylend-api/errs"
	"github.com/czConstant/constant-nftylend-api/models"
	"github.com/czConstant/constant-nftylend-api/serializers"
	"github.com/gin-gonic/gin"
)

func (s *Server) NftLendUpdateBlock(c *gin.Context) {
	ctx := s.requestContext(c)
	blockNumber, err := s.uint64FromContextParam(c, "block")
	if err != nil {
		ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	err = s.nls.LendNftLendUpdateBlock(ctx, blockNumber)
	if err != nil {
		ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	ctxJSON(c, http.StatusOK, &serializers.Resp{Result: true})
}

func (s *Server) BlockchainScanBlock(c *gin.Context) {
	ctx := s.requestContext(c)
	blockNumber, err := s.uint64FromContextParam(c, "block")
	if err != nil {
		ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	network := models.Network(c.Param("network"))
	switch network {
	case models.NetworkSOL:
		{
			err = s.nls.LendNftLendUpdateBlock(ctx, blockNumber)
			if err != nil {
				ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
				return
			}
		}
	case models.NetworkMATIC,
		models.NetworkETH:
		{
			err = s.nls.JobEvmNftypawnFilterLogs(ctx, network, blockNumber)
			if err != nil {
				ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
				return
			}
		}
	default:
		{
			ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(errs.ErrBadRequest)})
			return
		}
	}
	ctxJSON(c, http.StatusOK, &serializers.Resp{Result: true})
}

func (s *Server) LenInternalHookSolanaInstruction(c *gin.Context) {
	ctx := s.requestContext(c)
	var req struct {
		BlockNumber      uint64      `json:"block_number"`
		BlockTime        uint64      `json:"block_time"`
		TransactionHash  string      `json:"transaction_hash"`
		TransactionIndex uint        `json:"transaction_index"`
		InstructionIndex uint        `json:"instruction_index"`
		Program          string      `json:"program"`
		Instruction      string      `json:"instruction"`
		Data             interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ctxJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	err := s.nls.InternalHookSolanaInstruction(ctx, models.NetworkSOL, req.BlockNumber, req.BlockTime, req.TransactionHash, req.TransactionIndex, req.InstructionIndex, req.Program, req.Instruction, req.Data)
	if err != nil {
		ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	ctxJSON(c, http.StatusOK, &serializers.Resp{Result: true})
}

func (s *Server) JobEvmNftypawnFilterLogs(c *gin.Context) {
	ctx := s.requestContext(c)
	err := s.nls.JobEvmNftypawnFilterLogs(ctx, models.NetworkMATIC, 0)
	if err != nil {
		ctxAbortWithStatusJSON(c, http.StatusBadRequest, &serializers.Resp{Error: errs.NewError(err)})
		return
	}
	ctxJSON(c, http.StatusOK, &serializers.Resp{Result: true})
}