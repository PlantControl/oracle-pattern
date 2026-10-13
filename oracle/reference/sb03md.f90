program sb03md_oracle
  implicit none
  character :: dico, job, fact, trans
  integer :: n, pad, bad, nn, la, lu, lc, lw, info, i, j, stat, handler
  integer, allocatable :: iw(:)
  real(8), allocatable :: a(:,:), u(:,:), c(:,:), wr(:), wi(:), work(:)
  real(8) :: scale, sep, ferr, true_sep
  real(8), parameter :: guard = -987654321.125d0
  logical :: padding_ok
  common /error_observation/ handler
  read(*,*,iostat=stat) dico, job, fact, trans, n, pad, bad
  if (stat /= 0 .or. n < 0 .or. pad < 0) stop 2
  la=max(1,n)+pad; lu=la; lc=la; lw=max(1,2*n*n+100*n)
  allocate(a(la,max(1,n)),u(lu,max(1,n)),c(lc,max(1,n)),wr(n),wi(n),iw(max(1,n*n)),work(lw))
  a=guard; u=guard; c=guard; wr=guard; wi=guard; work=guard
  scale=guard; sep=guard; ferr=guard
  do i=1,n
    read(*,*) (a(i,j),j=1,n)
  end do
  do i=1,n
    read(*,*) (u(i,j),j=1,n)
  end do
  do i=1,n
    read(*,*) (c(i,j),j=1,n)
  end do
  nn=n
  select case(bad)
  case(1)
    dico='?'
  case(2)
    job='?'
  case(3)
    fact='?'
  case(4)
    trans='?'
  case(5)
    nn=-1
  case(7)
    la=0
  case(9)
    lu=0
  case(11)
    lc=0
  case(19)
    lw=0
  end select
  true_sep=guard
  if(bad==0.and.(fact=='N'.or.fact=='n')) call separation_svd(n,dico,a,size(a,1),true_sep)
  handler=0
  call SB03MD(dico,job,fact,trans,nn,a,la,u,lu,c,lc,scale,sep,ferr,wr,wi,iw,work,lw,info)
  padding_ok=all(a(n+1:,:)==guard).and.all(u(n+1:,:)==guard).and.all(c(n+1:,:)==guard)
  if(n==0) padding_ok=padding_ok.and.all(a==guard).and.all(u==guard).and.all(c==guard)
  write(*,*) info,handler,merge(1,0,padding_ok)
  write(*,'(*(ES26.17E3,1X))') scale,sep,ferr
  do i=1,n
    write(*,'(*(ES26.17E3,1X))') (a(i,j),j=1,n)
  end do
  do i=1,n
    write(*,'(*(ES26.17E3,1X))') (u(i,j),j=1,n)
  end do
  do i=1,n
    write(*,'(*(ES26.17E3,1X))') (c(i,j),j=1,n)
  end do
  write(*,'(*(ES26.17E3,1X))') wr
  write(*,'(*(ES26.17E3,1X))') wi
  write(*,'(*(ES26.17E3,1X))') true_sep
end program
subroutine xerbla(name,argument)
  implicit none
  character(*) :: name
  integer :: argument,handler
  common /error_observation/ handler
  if(name/='SB03MD') stop 3
  handler=argument
end subroutine
logical function select(wr,wi)
  implicit none
  real(8) :: wr,wi
  select=.false.
end function

subroutine separation_svd(n,dico,a,lda,sep)
  implicit none
  integer :: n,lda,m,i,j,k,l,row,col,info,lw
  character :: dico
  real(8) :: a(lda,*),sep,dummy(1)
  real(8),allocatable :: op(:,:),sv(:),work(:)
  m=n*n
  sep=0d0
  if(m==0) return
  lw=max(1,10*m)
  allocate(op(m,m),sv(m),work(lw))
  do j=1,n
    do i=1,n
      row=i+(j-1)*n
      do l=1,n
        do k=1,n
          col=k+(l-1)*n
          if(dico=='C'.or.dico=='c') then
            op(row,col)=0d0
            if(j==l) op(row,col)=op(row,col)+a(k,i)
            if(i==k) op(row,col)=op(row,col)+a(l,j)
          else
            op(row,col)=a(k,i)*a(l,j)
            if(i==k.and.j==l) op(row,col)=op(row,col)-1d0
          end if
        end do
      end do
    end do
  end do
  call DGESVD('N','N',m,m,op,m,sv,dummy,1,dummy,1,work,lw,info)
  if(info/=0) stop 4
  sep=sv(m)
end subroutine
