using System;
using System.Collections.Generic;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        bool SimulateElasticSearchGet();
        void SimulateElasticSearchSet(int x);
    }
}
